package spark

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodCallback interface {
	OnAddRunningPod(pod *v1.Pod)
	OnUpdateStatusPod(pod *v1.Pod)
}

func Watch(clientset *kubernetes.Clientset, namespace string, podCB PodCallback, stopCh <-chan struct{}) {
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), string(v1.ResourcePods), namespace,
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				newPod, newOk := newObj.(*v1.Pod)
				oldPod, oldOk := oldObj.(*v1.Pod)
				if newOk && oldOk {
					// Filter spark driver pod
					if role, ok := newPod.Labels["spark-role"]; ok && role == "driver" {
						// Trigger when status changed
						if newPod.Status.Phase != oldPod.Status.Phase {
							if newPod.Status.Phase == v1.PodRunning {
								podCB.OnAddRunningPod(newPod)
							} else {
								podCB.OnUpdateStatusPod(newPod)
							}
						}
					}
				}
			},
		},
	)

	controller.Run(stopCh)
}
