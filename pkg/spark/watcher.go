package spark

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/fields"
	"time"
	"k8s.io/client-go/tools/cache"
)

type PodCallback interface {
	OnAddRunningPod(pod *v1.Pod)
	OnUpdatePod(oldPod *v1.Pod, newPod *v1.Pod)
}

func Watch(clientset *kubernetes.Clientset, namespace string, podCB PodCallback, stopCh <-chan struct{}) {
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), string(v1.ResourcePods), namespace,
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if ok {
					if role, ok := pod.Labels["spark-role"]; ok && role == "driver" {
						glog.Infof("wait pod %v for running.", pod.Name)
						go waitPodRunning(clientset, pod, podCB.OnAddRunningPod)
					}
				} else {
					glog.Errorf("fail to receive add of pod %v.", obj)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newPod, newOk := newObj.(*v1.Pod)
				oldPod, oldOk := oldObj.(*v1.Pod)
				if newOk && oldOk {
					podCB.OnUpdatePod(oldPod, newPod)
				} else {
					glog.Errorf("fail to receive update of pod %v.", oldObj)
				}
			},
		},
	)

	controller.Run(stopCh)
}

func waitPodRunning(clientset *kubernetes.Clientset, pod *v1.Pod, callback func(pod *v1.Pod)) {
	switch status := pod.Status.Phase; status {
	case v1.PodRunning:
		glog.Infof("pod %v is running.", pod.Name)
		callback(pod)
	case v1.PodPending:
		doCheckRunning(clientset, pod, callback)
	}
}

func doCheckRunning(clientset *kubernetes.Clientset, pod *v1.Pod, callback func(pod *v1.Pod)) {
	retry := 10
	for {
		pods, err := clientset.CoreV1().Pods(pod.Namespace).List(metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("metadata.name", pod.Name).String()})
		if err != nil {
			glog.Errorf("fail to get pod: %v", err)
			return
		}
		if pods != nil && len(pods.Items) >= 1 {
			pod := pods.Items[0]
			if pod.Status.Phase == v1.PodPending {
				retry--
				if retry == 0 {
					glog.Infof("fail to retrieve pod %v because of exceed retry time.", pod.Name)
					return
				}
				time.Sleep(2 * time.Second)
				continue
			}
			if pod.Status.Phase == v1.PodRunning {
				glog.Infof("pod %v is running.", pod.Name)
				callback(&pod)
				return
			}
		} else {
			glog.Errorf("cannot find pod: %v", pod.Name)
			return
		}
	}

}
