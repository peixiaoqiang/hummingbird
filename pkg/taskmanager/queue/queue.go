package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/TalkingData/hummingbird/pkg/taskmanager/model"
	v3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

var (
	ErrKeyExists      = errors.New("key already exists")
	ErrWaitMismatch   = errors.New("unexpected wait result")
	ErrTooManyClients = errors.New("too many clients")
	ErrNoWatcher      = errors.New("no watcher channel")
)

type TaskQueue interface {
	Enqueue(*model.Task) error
	Dequeue() (*model.Task, error)
	Get(string) (*model.Task, error)
	GetFirst() (*model.Task, error)
	Remove(string) error
}

// ETCDQueue implements a multi-reader, multi-writer distributed queue.
type ETCDQueue struct {
	client *v3.Client
	ctx    context.Context

	keyPrefix string
}

func (q *ETCDQueue) encode(task model.Task) ([]byte, error) {
	data, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (q *ETCDQueue) decode(data []byte) (*model.Task, error) {
	task := new(model.Task)
	err := json.Unmarshal(data, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func NewQueue(conf v3.Config, keyPrefix string) (*ETCDQueue, error) {
	client, err := v3.New(conf)
	if err != nil {
		return nil, err
	}
	return &ETCDQueue{client, context.TODO(), keyPrefix}, nil
}

func (q *ETCDQueue) Get(key string) (*model.Task, error) {
	resp, err := q.client.Get(q.ctx, fmt.Sprintf("%s/%s", q.keyPrefix, key))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) > 0 {
		return q.decode(resp.Kvs[0].Value)
	}

	return nil, nil
}

func (q *ETCDQueue) GetFirst() (*model.Task, error) {
	esp, err := q.client.Get(q.ctx, q.keyPrefix, v3.WithFirstRev()...)
	if err != nil {
		return nil, err
	}
	if len(esp.Kvs) > 0 {
		return q.decode(esp.Kvs[0].Value)
	}
	return nil, nil
}

func (q *ETCDQueue) Remove(key string) error {
	_, err := q.client.Delete(q.ctx, path.Join(q.keyPrefix, key))
	if err != nil {
		return err
	}
	return nil
}

func (q *ETCDQueue) Enqueue(task *model.Task) error {
	data, err := q.encode(*task)
	if err != nil {
		return err
	}
	_, err = newKV(q.client, fmt.Sprintf("%s/%s", q.keyPrefix, task.ID), string(data), v3.NoLease)
	return err
}

// Dequeue returns Enqueue()'d elements in FIFO order. If the
// queue is empty, Dequeue blocks until elements are available.
func (q *ETCDQueue) Dequeue() (*model.Task, error) {
	// TODO: fewer round trips by fetching more than one key
	resp, err := q.client.Get(q.ctx, q.keyPrefix, v3.WithFirstRev()...)
	if err != nil {
		return nil, err
	}

	kv, err := claimFirstKey(q.client, resp.Kvs)
	if err != nil {
		return nil, err
	} else if kv != nil {
		task := new(model.Task)
		err := json.Unmarshal(kv.Value, task)
		if err != nil {
			return nil, err
		}
		return task, nil
	} else if resp.More {
		// missed some items, retry to read in more
		return q.Dequeue()
	}

	// nothing yet; wait on elements
	ev, err := WaitPrefixEvents(
		q.client,
		q.keyPrefix,
		resp.Header.Revision,
		[]mvccpb.Event_EventType{mvccpb.PUT})
	if err != nil {
		return nil, err
	}

	ok, err := deleteRevKey(q.client, string(ev.Kv.Key), ev.Kv.ModRevision)
	if err != nil {
		return nil, err
	} else if !ok {
		return q.Dequeue()
	}

	task := new(model.Task)
	err = json.Unmarshal(ev.Kv.Value, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// deleteRevKey deletes a key by revision, returning false if key is missing
func deleteRevKey(kv v3.KV, key string, rev int64) (bool, error) {
	cmp := v3.Compare(v3.ModRevision(key), "=", rev)
	req := v3.OpDelete(key)
	txnresp, err := kv.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return false, err
	} else if !txnresp.Succeeded {
		return false, nil
	}
	return true, nil
}

func claimFirstKey(kv v3.KV, kvs []*mvccpb.KeyValue) (*mvccpb.KeyValue, error) {
	for _, k := range kvs {
		ok, err := deleteRevKey(kv, string(k.Key), k.ModRevision)
		if err != nil {
			return nil, err
		} else if ok {
			return k, nil
		}
	}
	return nil, nil
}
