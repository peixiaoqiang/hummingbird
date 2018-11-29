package taskmanager

import (
	"fmt"
	"path"
	"time"

	"github.com/golang/glog"

	"github.com/TalkingData/hummingbird/pkg/taskmanager/model"
	"github.com/TalkingData/hummingbird/pkg/taskmanager/queue"
	v3 "go.etcd.io/etcd/clientv3"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

type ResourcePool interface {
	Allocate(*model.Task) error
	Init(string, *ResourceQuota) error
}

type KubernetesResourceQuota struct {
	client *k8s.Clientset
	conf   *Conf
}

func (c *KubernetesResourceQuota) Fit(task *model.Task) bool {
	quota, err := c.client.CoreV1().ResourceQuotas(c.conf.Namespace).Get(c.conf.ResourceQuotaName, meta_v1.GetOptions{})
	if err != nil {
		return false
	}
	hard := quota.Status.Hard
	used := quota.Status.Used
	freeCPU := hard[v1.ResourceRequestsCPU].DeepCopy()
	usedCPU := used[v1.ResourceRequestsCPU]
	freeCPU.Sub(usedCPU)
	freeMem := hard[v1.ResourceRequestsMemory].DeepCopy()
	usedMem := used[v1.ResourceRequestsMemory]
	freeMem.Sub(usedMem)

	totalCPU, totalMem := task.TotalResource()
	return freeCPU.CmpInt64(totalCPU) >= 0 && freeMem.CmpInt64(totalMem<<20) >= 0
}

type TaskManager interface {
	Run(model.Task, chan<- struct{}) error
	Pick(chan<- struct{}) error
	// Get(string) (Task, error)
	// Delete(string) error
	// List() ([]Task, error)
}

type TaskManagerImpl struct {
	waiting      queue.TaskQueue
	running      queue.TaskQueue
	failed       queue.TaskQueue
	finish       queue.TaskQueue
	taskEngine   TaskEngine
	resourcePool ResourcePool
}

func newTaskQueue(conf *Conf, name string) (queue.TaskQueue, error) {
	config := v3.Config{Endpoints: conf.EtcdIps}
	return queue.NewQueue(config, path.Join(conf.StorageKeyPrefix, "queue", name))
}

func NewTaskManager(client *k8s.Clientset, conf *Conf) (*TaskManagerImpl, error) {
	waiting, err := newTaskQueue(conf, "waiting")
	if err != nil {
		return nil, err
	}
	running, err := newTaskQueue(conf, "running")

	if err != nil {
		return nil, err
	}
	failed, err := newTaskQueue(conf, "failed")
	if err != nil {
		return nil, err
	}
	finish, err := newTaskQueue(conf, "finish")
	if err != nil {
		return nil, err
	}
	taskEngine := &SparkOnK8SEngine{Clientset: client, DeployMode: "cluster", Image: conf.Image, MasterIP: conf.MasterIP, ServiceAccoutName: conf.ServiceAccountName, NameSpace: conf.Namespace}
	resourcePool, err := NewResourceQuotaManager(conf)
	if err != nil {
		return nil, err
	}
	err = resourcePool.Init(conf.ResourceQuotaName, &conf.ResourceQuota)
	if err != nil {
		return nil, err
	}
	return &TaskManagerImpl{resourcePool: resourcePool, waiting: waiting, running: running, failed: failed, finish: finish, taskEngine: taskEngine}, nil
}

func (t *TaskManagerImpl) Run(task model.Task, finish chan<- struct{}) error {
	task.CreateTime = time.Now().String()
	err := t.resourcePool.Allocate(&task)
	if err != nil {
		if finish != nil {
			finish <- struct{}{}
		}
		return t.waiting.Enqueue(&task)
	}
	task.StartTime = time.Now().String()
	if err := t.running.Enqueue(&task); err != nil {
		if finish != nil {
			finish <- struct{}{}
		}
		return err
	}
	glog.V(1).Infof("start to run task: %v", task)
	t.run(task, finish)
	return nil
}

func (t *TaskManagerImpl) Pick(finish chan<- struct{}) error {
	task, err := t.waiting.GetFirst()
	glog.V(1).Infof("get first task: %v", task)
	if err != nil {
		return err
	}
	return t.Run(*task, finish)
}

func (t *TaskManagerImpl) run(task model.Task, finish chan<- struct{}) {
	go func() {
		if err := t.taskEngine.Submit(task); err != nil {
			t.running.Remove(task.ID)
			task.EndTime = time.Now().String()
			task.FailedReason = fmt.Sprintf("%v", err)
			t.failed.Enqueue(&task)
			if finish != nil {
				finish <- struct{}{}
			}
			return
		}
		t.running.Remove(task.ID)
		state := t.taskEngine.GetTaskState(&task)
		task.EndTime = time.Now().String()
		switch state {
		case Failed:
			glog.V(1).Infof("task failed: %v", task)
			t.failed.Enqueue(&task)
		case Finished:
			glog.V(1).Infof("task finished: %v", task)
			t.finish.Enqueue(&task)
		}
		if finish != nil {
			finish <- struct{}{}
		}
	}()
}
