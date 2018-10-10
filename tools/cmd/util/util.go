package util

import (
	kube "github.com/TalkingData/hummingbird/pkg/kubernetes"
	"k8s.io/client-go/kubernetes"
)

func GetKubeClient(kubeConf string) (client *kubernetes.Clientset, err error) {
	if kubeConf != "" {
		client, err = kube.GetClient(false, kubeConf)
	} else {
		client, err = kube.GetClient(true, "")
	}
	return
}
