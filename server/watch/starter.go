package main

import (
	"flag"
	"os"
	"encoding/json"
	"io/ioutil"
	"github.com/golang/glog"
	"github.com/TalkingData/hummingbird/pkg/kubernetes"
	"github.com/TalkingData/hummingbird/pkg/spark"
	"path"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend/factory"
)

var (
	conf = flag.String("conf", "/etc/hummingbird/spark.conf", "The configuration path")
)

type Conf struct {
	Namespace   string   `json:"namespace,omitempty"`
	SparkUIPort int      `json:"spark_ui_port,omitempty"`
	Kubeconfig  string   `json:"kubeconfig,omitempty"`
	EtcdIps     []string `json:"etcd_ips,omitempty"`
}

var CONF = &Conf{
	Namespace:   "default",
	SparkUIPort: 4040,
	Kubeconfig:  path.Join(homeDir(), ".kube", "config"),
	EtcdIps:     []string{"http://localhost:2379"},
}

func initConfig(configPath string) error {
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

func main() {
	flag.Parse()
	defer glog.Flush()

	initConfig(*conf)

	clientset, err := kubernetes.GetClient(CONF.Kubeconfig)
	if err != nil {
		glog.Fatalf("fail to get kubernetes client: %v", err)
	}

	sparkHandler := spark.ApplicationHandler{
		Namespace:         CONF.Namespace,
		SparkUIPort:       CONF.SparkUIPort,
		StoragePathPrefix: "/spark",
	}

	storeConfig := &storagebackend.Config{Type: storagebackend.StorageTypeETCD2, ServerList: CONF.EtcdIps}
	store, _ := factory.NewRawStorage(storeConfig)
	sparkHandler.Storage = store

	spark.Watch(clientset, CONF.Namespace, &sparkHandler)
}
