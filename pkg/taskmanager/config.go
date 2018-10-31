package taskmanager

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

type Conf struct {
	Namespace          string        `json:"namespace,omitempty"`
	ServiceAccountName string        `json:"service_account_name,omitempty"`
	ResourceQuota      ResourceQuota `json:"resource_quota,omitempty"`
	EtcdIps            []string      `json:"etcd_ips,omitempty"`
	StorageKeyPrefix   string        `json:"storage_key_prefix,omitempty"`
	K8SInClusterConfig bool          `json:"k8s_incluster_config,omitempty"`
	Kubeconfig         string        `json:"kubeconfig,omitempty"`
	MasterIP           string        `json:"master_ip,omitempty"`
	Image              string        `json:"image,omitempty"`
	ResourceQuotaName  string        `json:"resource_quota_name,omitempty"`
	SyncInterval       int           `json:"sync_interval,omitempty"`
}

var CONF = &Conf{
	Kubeconfig:         path.Join(homeDir(), ".kube", "config"),
	EtcdIps:            []string{"http://localhost:2379"},
	StorageKeyPrefix:   "/taskmanager/",
	K8SInClusterConfig: false,
	SyncInterval:       5,
}

func InitConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, CONF)
	if err != nil {
		return err
	}
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
