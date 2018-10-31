package taskmanager

import (
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func Run(clientset *kubernetes.Clientset, namespace string, taskManager TaskManager, stopCh <-chan struct{}) {
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), string(v1.ResourcePods), namespace,
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if ok {
					glog.V(1).Infof("catch deleted pod: %v", pod.Name)
					if err := taskManager.Pick(nil); err != nil {
						glog.Errorf("fail to pick task: %v", err)
					}
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				pod, ok := newObj.(*v1.Pod)
				if ok {
					if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
						glog.V(1).Infof("catch succeeded and failed pod: %v", pod.Name)
						if err := taskManager.Pick(nil); err != nil {
							glog.Errorf("fail to pick task: %v", err)
						}
					}
				}
			},
		},
	)

	controller.Run(stopCh)
}

func Sync(interval int, taskManager TaskManager, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			glog.Infoln("stop task manager syncing")
			return
		default:
			break
		}

		if err := taskManager.Pick(nil); err != nil {
			glog.Errorf("fail to pick task %v", err)
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
