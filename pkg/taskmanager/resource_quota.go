package taskmanager

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/TalkingData/hummingbird/pkg/storage"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend/factory"
	"github.com/TalkingData/hummingbird/pkg/taskmanager/model"
)

type ResourceQuota struct {
	LimitCPU int64 `json:"limit_cpu,omitempty"`
	LimitMem int64 `json:"limit_memory,omitempty"`
	UsedCPU  int64
	UsedMem  int64
}

type ResourceAllocation struct {
	CPU    int64
	Memory int64
}

type ResourceQuotaManager struct {
	lock    sync.Mutex
	storage storage.Interface
	conf    *Conf
	context context.Context
}

func NewResourceQuotaManager(conf *Conf) (*ResourceQuotaManager, error) {
	storeConfig := &storagebackend.Config{Type: storagebackend.StorageTypeETCD3, ServerList: CONF.EtcdIps, Prefix: path.Join(conf.StorageKeyPrefix, "resourcequota")}
	store, _ := factory.NewRawStorage(storeConfig)
	return &ResourceQuotaManager{storage: store, conf: conf, context: context.TODO()}, nil

}

func (r *ResourceQuotaManager) getCurrentQuota() (*ResourceQuota, error) {
	res := new(ResourceQuota)
	err := r.storage.Get(r.context, r.conf.ResourceQuotaName, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ResourceQuotaManager) Allocate(task *model.Task) error {
	// TODO need a rw lock when using cluster
	r.lock.Lock()
	defer r.lock.Unlock()
	resQuota, err := r.getCurrentQuota()
	if err != nil {
		return err
	}
	totalCPU, totalMem := task.TotalResource()
	remainCPU := resQuota.LimitCPU - resQuota.UsedCPU - totalCPU
	remainMem := resQuota.LimitMem - resQuota.UsedMem - totalMem
	if remainCPU >= 0 && remainMem >= 0 {
		resQuota.UsedCPU = resQuota.UsedCPU + totalCPU
		resQuota.UsedMem = resQuota.UsedMem + totalMem
		err := r.storage.Update(r.context, r.conf.ResourceQuotaName, resQuota)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("exceeded resource quota, task: %+v, resource quota: %+v", task, resQuota)
}

func (r *ResourceQuotaManager) Init(name string, quota *ResourceQuota) error {
	err := r.storage.Create(r.context, name, quota)
	if err == storage.ErrKeyExists {
		return nil
	}
	return err
}
