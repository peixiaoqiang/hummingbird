package spark

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/TalkingData/hummingbird/pkg/storage"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

const SparkDriverName string = "spark-kubernetes-driver"

type Job struct {
	ID                  int    `json:"jobId"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	NumActiveTasks      int    `json:"numActiveTasks"`
	NumTasks            int    `json:"numTasks"`
	NumCompletedTasks   int    `json:"numCompletedTasks"`
	NumSkippedTasks     int    `json:"numSkippedTasks"`
	NumFailedTasks      int    `json:"numFailedTasks"`
	NumKilledTasks      int    `json:"numKilledTasks"`
	NumCompletedIndices int    `json:"numCompletedIndices"`
	NumActiveStages     int    `json:"numActiveStages"`
	NumCompletedStages  int    `json:"numCompletedStages"`
	NumSkippedStages    int    `json:"numSkippedStages"`
	NumFailedStages     int    `json:"numFailedStages"`
}

type AppAttempt struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type SparkApplication struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Attempts []AppAttempt `json:"attempts"`
}

type Application struct {
	ID               string       `json:"id"`
	Attempts         []AppAttempt `json:"attempts"`
	Name             string       `json:"name"`
	DriverIP         string       `json:"driver_ip"`
	HistoryServerURL string       `json:"history_server_url"`
	StartTime        string       `json:"start_time"`
	EndTime          string       `json:"end_time"`
	Status           string       `json:"status"`
	Jobs             []Job        `json:"jobs"`
	DriverStatus     string       `json:"driver_status"`
}

func (handler *ApplicationHandler) getSparkApplicationID(driverIP string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/api/v1/applications", driverIP, handler.SparkUIPort))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	apps := []SparkApplication{}
	err = json.Unmarshal(body, &apps)
	if len(apps) > 0 {
		return apps[0].ID, nil
	}
	return "", nil
}

func (handler *ApplicationHandler) getSparkJobs(url string) ([]Job, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	jobs := []Job{}
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (handler *ApplicationHandler) getSparkApplication(url string) (*SparkApplication, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	app := SparkApplication{}
	err = json.Unmarshal(body, &app)
	if err != nil {
		return nil, err
	}
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
	sparkApp, err := handler.getSparkApplication(app.HistoryServerURL)
	if err != nil {
		glog.Errorf("fail to get spark application %v, error is %v.", app.ID, err)
	} else {
		app.Attempts = sparkApp.Attempts
	}
	app.Jobs, err = handler.getSparkJobs(fmt.Sprintf("%v/jobs", app.HistoryServerURL))
	if err != nil {
		glog.Errorf("fail to get spark %v jobs, error is %v.", app.Name, err)
	}
	return app, nil
}

func (handler *ApplicationHandler) OnAddRunningPod(pod *v1.Pod) {
	glog.Infof("pod %v status is changing to %v.", pod.Name, string(pod.Status.Phase))
	retry := 5
	for {
		sparkAppID, err := handler.getSparkApplicationID(pod.Status.PodIP)
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
			newApp.ID = sparkAppID
			newApp.Status = string(v1.PodRunning)
			newApp.HistoryServerURL = fmt.Sprintf("http://%s:%d/api/v1/applications/%v", newApp.DriverIP, handler.SparkUIPort, newApp.ID)
			newApp.StartTime = time.Now().Format(time.RFC3339)
			err = handler.Storage.Create(context.TODO(), path.Join(handler.StoragePathPrefix, "applications", pod.Name), newApp)
			if err != nil {
				glog.Errorf("fail to update spark application, the pod is %v, erorr is %v.", pod.Name, err)
			}
			return
		}
	}
}

func (handler *ApplicationHandler) OnUpdateStatusPod(pod *v1.Pod) {
	glog.Infof("pod %v status is changing to %v.", pod.Name, string(pod.Status.Phase))
	app := &Application{}
	err := handler.Storage.Get(context.TODO(), path.Join(handler.StoragePathPrefix, "applications", pod.Name), app)
	if err != nil {
		glog.Errorf("fail to get spark application %v, error is %v.", pod.Name, err)
		return
	}
	app.HistoryServerURL = fmt.Sprintf("http://%v/api/v1/applications/%v", handler.SparkHistoryURL, app.ID)
	app.Status = string(pod.Status.Phase)
	if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
		app.EndTime = time.Now().Format(time.RFC3339)
	}
	for _, s := range pod.Status.ContainerStatuses {
		if s.Name == SparkDriverName {
			state := s.State
			if state.Waiting != nil {
				app.DriverStatus = "Waiting"
			} else if state.Running != nil {
				app.DriverStatus = "Running"
			} else if state.Terminated != nil {
				if state.Terminated.ExitCode == 0 {
					app.DriverStatus = "Succeeded"
				} else {
					app.DriverStatus = "Failed"
				}
			}
			break
		}
	}
	err = handler.Storage.CreateOrUpdate(context.TODO(), path.Join(handler.StoragePathPrefix, "applications", pod.Name), app)
	if err != nil {
		glog.Errorf("fail to update application %v, error is %v.", pod.Name, err)
	}
}
