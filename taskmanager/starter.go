package main

import (
	"flag"
	"sync"

	"github.com/TalkingData/hummingbird/pkg/kubernetes"
	"github.com/TalkingData/hummingbird/pkg/taskmanager"
	"github.com/golang/glog"
	k8s "k8s.io/client-go/kubernetes"
)

var (
	conf      = flag.String("conf", "/etc/hummingbird/taskmanager.conf", "The configuration path")
	clientset *k8s.Clientset
)

func main() {
	flag.Parse()
	defer glog.Flush()

	taskmanager.InitConfig(*conf)

	clientset, err := kubernetes.GetClient(taskmanager.CONF.K8SInClusterConfig, taskmanager.CONF.Kubeconfig)
	if err != nil {
		glog.Fatalf("fail to get kubernetes client: %v", err)
	}

	stop := make(chan struct{})
	defer close(stop)

	tm, err := taskmanager.NewTaskManager(clientset, taskmanager.CONF)
	if err != nil {
		glog.Fatalf("fail to start task manager: %v", err)
	}
	glog.Info("start task manager")
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		// run task controller
		taskmanager.Run(clientset, taskmanager.CONF.Namespace, tm, stop)
	}()

	go func() {
		defer wg.Done()
		// run task sync
		taskmanager.Sync(taskmanager.CONF.SyncInterval, tm, stop)
	}()

	wg.Wait()
	glog.Infoln("task manager daemon has stopped")
}
