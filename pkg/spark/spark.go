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

type Job struct {
	ID                  int    `json:"jobId,omitempty"`
	Name                string `json:"name,omitempty"`
	Status              string `json:"status,omitempty"`
	NumActiveTasks      int    `json:"numActiveTasks,omitempty"`
	NumTasks            int    `json:"numTasks,omitempty"`
	NumCompletedTasks   int    `json:"numCompletedTasks,omitempty"`
	NumSkippedTasks     int    `json:"numSkippedTasks,omitempty"`
	NumFailedTasks      int    `json:"numFailedTasks,omitempty"`
	NumKilledTasks      int    `json:"numKilledTasks,omitempty"`
	NumCompletedIndices int    `json:"numCompletedIndices,omitempty"`
	NumActiveStages     int    `json:"numActiveStages,omitempty"`
	NumCompletedStages  int    `json:"numCompletedStages,omitempty"`
	NumSkippedStages    int    `json:"numSkippedStages,omitempty"`
	NumFailedStages     int    `json:"numFailedStages,omitempty"`
	killedTasksSummary  int    `json:"killedTasksSummary,omitempty"`
}

type AppAttempt struct {
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

type Application struct {
	ID       string       `json:"id,omitempty"`
	Name     string       `json:"name,omitempty"`
	DriverIP string       `json:"driver_ip,omitempty"`
	Jobs     []Job        `json:"jobs,omitempty"`
	Attempts []AppAttempt `json:"attempts,omitempty"`
}

func (handler *ApplicationHandler) getApplicationID(driverIP string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/api/v1/applications", driverIP, handler.SparkUIPort))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	apps := []Application{}
	err = json.Unmarshal(body, &apps)
	if len(apps) > 0 {
		return apps[0].ID, nil
	}
	return "", nil
}

func (handler *ApplicationHandler) getJobs(url string) ([]Job, error) {
	resp, err := http.Get(url)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	jobs := []Job{}
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	return jobs, nil
}

func (handler *ApplicationHandler) getApplication(url string) (*Application, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	app := Application{}
	err = json.Unmarshal(body, &app)
	if err != nil {
		return nil, err
	}
	app.Jobs, _ = handler.getJobs(fmt.Sprintf("%v/jobs", url))
	return &app, nil
}

type ApplicationHandler struct {
	Storage           storage.Interface
	StoragePathPrefix string
	Namespace         string
	SparkUIPort       int
	SparkHistoryURL   string
}

func (handler *ApplicationHandler) GetApplication(appName string) (*Application, error) {
	app := &Application{}
	err := handler.Storage.Get(context.TODO(), path.Join(handler.StoragePathPrefix, "applications", appName), app)
	if err != nil {
		return nil, err
	}
	status := ""
	handler.Storage.Get(context.TODO(), path.Join(handler.StoragePathPrefix, "status", appName), &status)
	url := fmt.Sprintf("http://%v/api/v1/applications/%v", handler.SparkHistoryURL, app.ID)
	// If pod is running, it will retrieve application from spark history server
	if status == string(v1.PodRunning) {
		url = fmt.Sprintf("http://%s:%d/api/v1/applications/%v", app.DriverIP, handler.SparkUIPort, app.ID)
	}
	app, err = handler.getApplication(url)
	if err != nil {
		return nil, err
	}
	app.Name = appName
	return app, nil
}

func (handler *ApplicationHandler) OnAddRunningPod(pod *v1.Pod) {
	retry := 5
	for {
		appID, err := handler.getApplicationID(pod.Status.PodIP)
		if err != nil {
			retry--
			if retry == 0 {
				glog.Errorf("fail to get spark application, the pod is %v, error is %v.", pod.Name, err)
				return
			}
			time.Sleep(2 * time.Second)
			continue
		} else {
			newApp := &Application{}
			newApp.DriverIP = pod.Status.PodIP
			newApp.Name = pod.Name
			newApp.ID = appID
			err = handler.Storage.Create(context.TODO(), path.Join(handler.StoragePathPrefix, "applications", pod.Name), newApp)
			if err != nil {
				glog.Errorf("fail to update spark application, the pod is %v, erorr is %v.", pod.Name, err)
			}
			return
		}
	}
}

func (handler *ApplicationHandler) GetApplicationStatus(appName string) (string, error) {
	status := ""
	err := handler.Storage.Get(context.TODO(), path.Join(handler.StoragePathPrefix, "status", appName), &status)
	if err != nil {
		return "", err
	}
	return status, nil
}

func (handler *ApplicationHandler) OnUpdatePod(oldPod *v1.Pod, newPod *v1.Pod) {
	glog.Infof("update pod %v, status is %v.", oldPod.Name, newPod.Status.Phase)
	handler.Storage.CreateOrUpdate(context.TODO(), path.Join(handler.StoragePathPrefix, "status", newPod.Name), newPod.Status.Phase)
}
