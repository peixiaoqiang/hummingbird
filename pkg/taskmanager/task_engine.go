package taskmanager

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/TalkingData/hummingbird/pkg/taskmanager/model"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const SparkDriverName string = "spark-kubernetes-driver"

const (
	Finished string = "Finished"
	Failed   string = "Failed"
	Unknown  string = "Unknown"
	Waiting  string = "Waiting"
	Running  string = "Running"
)

type TaskEngine interface {
	Submit(model.Task) error
	GetTaskState(*model.Task) string
}
type SparkOnK8SEngine struct {
	NameSpace         string
	DeployMode        string
	Image             string
	MasterIP          string
	ServiceAccoutName string
	Clientset         *k8s.Clientset
}

func (s *SparkOnK8SEngine) GetDriverName(task *model.Task) string {
	return fmt.Sprintf("%s-%s-driver", task.Name, task.ID)
}

func (s *SparkOnK8SEngine) GetTaskState(task *model.Task) string {
	pod, err := s.Clientset.CoreV1().Pods(s.NameSpace).Get(s.GetDriverName(task), meta_v1.GetOptions{})
	if err != nil {
		return Unknown
	}
	for _, s := range pod.Status.ContainerStatuses {
		if s.Name == SparkDriverName {
			state := s.State
			if state.Waiting != nil {
				return Waiting
			} else if state.Running != nil {
				return Running
			} else if state.Terminated != nil {
				if state.Terminated.ExitCode == 0 {
					return Finished
				} else {
					return Failed
				}
			}
		}
	}
	return Unknown
}

func (s *SparkOnK8SEngine) Submit(task model.Task) error {
	conf := make(map[string]string)
	conf["name"] = task.Name
	conf["deploy-mode"] = s.DeployMode
	if task.MasterIP != "" {
		conf["master"] = task.MasterIP
	} else {
		conf["master"] = s.MasterIP
	}
	conf["class"] = task.Class

	sparkConf := make(map[string]string)
	sparkConf["spark.driver.cores"] = strconv.FormatInt(task.DriverCPU, 10)
	sparkConf["spark.kubernetes.driver.limit.cores"] = strconv.FormatInt(task.DriverCPU, 10)
	sparkConf["spark.driver.memory"] = fmt.Sprintf("%sM", strconv.FormatInt(task.DriverMem, 10))
	sparkConf["spark.executor.cores"] = strconv.FormatInt(task.ExecutorCPU, 10)
	sparkConf["spark.executor.memory"] = fmt.Sprintf("%sM", strconv.FormatInt(task.ExecutorMem, 10))
	sparkConf["spark.executor.instances"] = strconv.FormatInt(task.ExecutorNum, 10)
	sparkConf["spark.kubernetes.driver.label.taskid"] = task.ID
	sparkConf["spark.kubernetes.executor.label.taskid"] = task.ID
	sparkConf["spark.kubernetes.driver.pod.name"] = s.GetDriverName(&task)
	if task.Image != "" {
		sparkConf["spark.kubernetes.container.image"] = task.Image
	} else {
		sparkConf["spark.kubernetes.container.image"] = s.Image
	}
	sparkConf["spark.kubernetes.authenticate.driver.serviceAccountName"] = s.ServiceAccoutName
	sparkConf["spark.kubernetes.namespace"] = s.NameSpace

	args := []string{}
	for k, v := range conf {
		args = append(args, fmt.Sprintf("--%s", k), v)
	}
	for k, v := range sparkConf {
		args = append(args, "--conf", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, task.Conf...)
	args = append(args, task.TaskArgs...)
	cmd := exec.Command("spark-submit", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("submit spark task: %v", stderr.String())
	}
	return nil
}
