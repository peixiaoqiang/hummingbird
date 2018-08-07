package main

import (
	"net/http"
	"flag"
	"strings"
	"os"
	"sync"
	"encoding/json"
	"io/ioutil"
	"github.com/golang/glog"
	"github.com/TalkingData/hummingbird/pkg/kubernetes"
	"github.com/TalkingData/hummingbird/pkg/spark"
	"path"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend/factory"
	"fmt"
)

var (
	conf = flag.String("conf", "/etc/hummingbird/spark.conf", "The configuration path")
)

type Conf struct {
	Namespace          string   `json:"namespace,omitempty"`
	SparkUIPort        int      `json:"spark_ui_port,omitempty"`
	Kubeconfig         string   `json:"kubeconfig,omitempty"`
	EtcdIps            []string `json:"etcd_ips,omitempty"`
	HttpPort           int      `json:"http_port,omitempty"`
	SparkHistoryURL    string   `json:"spark_history_url,omitempty"`
	StoragePrefix      string   `json:"storage_prefix,omitempty"`
	K8SInClusterConfig bool     `json:"k8s_incluster_config,omitempty"`
}

var CONF = &Conf{
	Namespace:     "default",
	SparkUIPort:   4040,
	Kubeconfig:    path.Join(homeDir(), ".kube", "config"),
	EtcdIps:       []string{"http://localhost:2379"},
	HttpPort:      9001,
	StoragePrefix: "/spark",
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

var SparkHandler *spark.ApplicationHandler
var waitgroup sync.WaitGroup

func main() {
	flag.Parse()
	defer glog.Flush()

	initConfig(*conf)

	clientset, err := kubernetes.GetClient(CONF.Kubeconfig)
	if err != nil {
		glog.Fatalf("fail to get kubernetes client: %v", err)
	}

	SparkHandler = &spark.ApplicationHandler{
		Namespace:         CONF.Namespace,
		SparkUIPort:       CONF.SparkUIPort,
		StoragePathPrefix: CONF.StoragePrefix,
		SparkHistoryURL:   CONF.SparkHistoryURL,
	}

	storeConfig := &storagebackend.Config{Type: storagebackend.StorageTypeETCD2, ServerList: CONF.EtcdIps}
	store, _ := factory.NewRawStorage(storeConfig)
	SparkHandler.Storage = store

	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		stop := make(chan struct{})
		defer close(stop)
		glog.Infoln("Start watching.")
		spark.Watch(clientset, CONF.Namespace, SparkHandler, stop)
	}()

	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		http.HandleFunc("/applications/", handleApplication)
		glog.Infof("Start server on %v", CONF.HttpPort)
		glog.Error(http.ListenAndServe(fmt.Sprintf(":%v", CONF.HttpPort), nil))
	}()

	waitgroup.Wait()
}

func handleApplication(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	i := strings.LastIndex(path, "/")
	appName := path[i+1:]
	glog.Infof("receive request of %v.", appName)
	app, err := SparkHandler.GetApplication(appName)
	if err != nil {
		glog.Errorf("fail to get spark application, application name is %v, error is %v.", appName, err)
		w.WriteHeader(500)
		return
	}
	appJson, err := json.Marshal(app)
	if err != nil {
		glog.Errorf("fail to serialize spark application, application name is %v, error is %v.", appName, err)
		w.WriteHeader(500)
		return
	} else {
		glog.Infof("response of %v is %v", appName, string(appJson))
		w.Write(appJson)
	}
}
