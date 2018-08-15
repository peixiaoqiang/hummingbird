package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/TalkingData/hummingbird/pkg/kubernetes"
	"github.com/TalkingData/hummingbird/pkg/spark"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend/factory"
	"github.com/golang/glog"
	etcd "github.com/coreos/etcd/client"
	"github.com/gorilla/mux"
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
	UIDir              string   `json:"ui_dir,omitempty"`
	UIPort             int      `json:"ui_port,omitempty"`
	EnableUI           bool     `json:"enable_ui,omitempty"`
}

var CONF = &Conf{
	Namespace:          "default",
	SparkUIPort:        4040,
	Kubeconfig:         path.Join(homeDir(), ".kube", "config"),
	EtcdIps:            []string{"http://localhost:2379"},
	HttpPort:           9001,
	StoragePrefix:      "/spark",
	K8SInClusterConfig: false,
	UIDir:              "/usr/local/spark-watch/html",
	UIPort:             80,
	EnableUI:           false,
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

	clientset, err := kubernetes.GetClient(CONF.K8SInClusterConfig, CONF.Kubeconfig)
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
		r := mux.NewRouter()
		r.Path("/applications/{name}").HandlerFunc(applicationHandler)
		r.Path("/applications/{name}").Queries("callback", "{.*}").HandlerFunc(applicationHandler)
		glog.Infof("Start server on %v", CONF.HttpPort)
		glog.Error(http.ListenAndServe(fmt.Sprintf(":%v", CONF.HttpPort), r))
	}()

	if CONF.EnableUI {
		waitgroup.Add(1)
		go func() {
			defer waitgroup.Done()
			glog.Infof("Start ui server on %v", CONF.UIPort)
			glog.Error(http.ListenAndServe(fmt.Sprintf(":%v", CONF.UIPort), http.FileServer(http.Dir(CONF.UIDir))))
		}()
	}

	waitgroup.Wait()
}

func applicationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["name"]
	glog.Infof("receive request of %v.", appName)
	app, err := SparkHandler.GetApplication(appName)
	if err != nil {
		if nerr, ok := err.(etcd.Error); ok && nerr.Code == etcd.ErrorCodeKeyNotFound {
			glog.Infof("can not found spark application %v.", appName)
			w.WriteHeader(404)
			return
		} else {
			glog.Errorf("fail to get spark application, application name is %v, error is %v.", appName, err)
			w.WriteHeader(400)
			return
		}
	}
	appJson, err := json.Marshal(app)
	if err != nil {
		glog.Errorf("fail to serialize spark application, application name is %v, error is %v.", appName, err)
		w.WriteHeader(400)
		return
	} else {
		glog.Infof("response of %v is %v", appName, string(appJson))
		callback, ok := r.URL.Query()["callback"]
		if ok {
			w.Write([]byte(fmt.Sprintf("%s(%s)", callback[0], appJson)))
		} else {
			w.Write(appJson)
		}
	}
}
