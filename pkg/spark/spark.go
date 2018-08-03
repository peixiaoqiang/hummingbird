package spark

import (
	"k8s.io/api/core/v1"
	"github.com/golang/glog"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"github.com/TalkingData/hummingbird/pkg/storage"
	"context"
	"path"
	"time"
)

type Application struct {
	ID   string
	Name string
}

func GetApplication(sparkDriverIP string, sparkUIPort int) (*Application, error) {
	driverURL := fmt.Sprintf("http://%s:%d/api/v1/applications", sparkDriverIP, sparkUIPort)
	glog.Infof("spark driver ui url is %v", driverURL)
	resp, err := http.Get(driverURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	apps := make([]Application, 0)
	err = json.Unmarshal(body, &apps)
	if err != nil {
		return nil, err
	}
	if len(apps) > 0 {
		return &apps[0], nil
	} else {
		return nil, nil
	}
}

type ApplicationHandler struct {
	Storage           storage.Interface
	StoragePathPrefix string
	Namespace         string
	SparkUIPort       int
}

func (handler *ApplicationHandler) OnAddRunningPod(pod *v1.Pod) {
	retry := 5
	for {
		app, err := GetApplication(pod.Status.PodIP, handler.SparkUIPort)
		if err != nil {
			retry--
			if retry == 0 {
				glog.Errorf("fail to get spark application, the pod is %v, error is %v.", pod.Name, err)
				return
			}
			time.Sleep(2 * time.Second)
			continue
		} else {
			err = handler.Storage.Create(context.TODO(), path.Join(handler.StoragePathPrefix, pod.Name), app.ID)
			if err != nil {
				glog.Errorf("fail to update spark application, the pod is %v, erorr is %v.", pod.Name, err)
			}
			return
		}
	}

}
