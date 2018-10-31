package taskmanager

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/TalkingData/hummingbird/pkg/kubernetes"
	"github.com/TalkingData/hummingbird/pkg/taskmanager/model"
	"github.com/google/uuid"
)

func clone(conf *Conf) *Conf {
	data, _ := json.Marshal(conf)
	newConfig := new(Conf)
	json.Unmarshal(data, newConfig)
	return newConfig
}

func testConfig() *Conf {
	conf := clone(CONF)
	conf.StorageKeyPrefix = "testtaskmanager"
	conf.EtcdIps = []string{"localhost:2379"}
	conf.Image = "172.20.0.112/library/spark_on_k8s:v1.5"
	conf.ServiceAccountName = "test-spark"
	conf.ResourceQuota = ResourceQuota{LimitCPU: 2, LimitMem: 1000}
	conf.Namespace = "test-spark"
	conf.MasterIP = "k8s://https://172.20.0.67:6443"
	conf.ResourceQuotaName = "test-spark"
	return conf
}

func TestAddTask(t *testing.T) {
	conf := testConfig()
	clientset, err := kubernetes.GetClient(false, conf.Kubeconfig)
	taskManager, _ := NewTaskManager(clientset, conf)

	id, _ := uuid.NewUUID()
	task := model.Task{
		ID:          id.String(),
		Name:        "test-task",
		DriverCPU:   1,
		DriverMem:   500,
		ExecutorCPU: 1,
		ExecutorMem: 500,
		ExecutorNum: 1,
		Class:       "org.apache.spark.examples.SparkPi",
		TaskArgs:    []string{"http://172.23.4.148/oam/jars/spark-examples_2.11-2.3.0.jar", "10"},
	}

	finish := make(chan struct{}, 1)
	defer close(finish)
	err = taskManager.Run(task, finish)
	if err != nil {
		t.Errorf("fail to start task: %v", err)
		return
	}
	<-finish
}

func TestPick(t *testing.T) {
	conf := testConfig()
	clientset, _ := kubernetes.GetClient(false, conf.Kubeconfig)
	taskManager, _ := NewTaskManager(clientset, conf)
	finish := make(chan struct{})
	err := taskManager.Pick(finish)
	if err != nil {
		t.Errorf("fail to pick task: %v", err)
		return
	}
	<-finish
}

func TestBatchAddTask(t *testing.T) {
	conf := testConfig()
	clientset, err := kubernetes.GetClient(false, conf.Kubeconfig)
	taskManager, _ := NewTaskManager(clientset, conf)

	num := 10
	wg := sync.WaitGroup{}
	wg.Add(num)
	for i := 0; i < num; i++ {
		id, _ := uuid.NewUUID()
		task := model.Task{
			ID:          id.String(),
			Name:        "test-task",
			DriverCPU:   1,
			DriverMem:   500,
			ExecutorCPU: 1,
			ExecutorMem: 500,
			ExecutorNum: 1,
			Class:       "org.apache.spark.examples.SparkPi",
			TaskArgs:    []string{"http://172.23.4.148/oam/jars/spark-examples_2.11-2.3.0.jar", "10"},
		}

		go func() {
			defer wg.Done()
			finish := make(chan struct{}, 1)
			defer close(finish)
			err = taskManager.Run(task, finish)
			if err != nil {
				t.Errorf("fail to start task: %v", err)
			}
			<-finish
		}()
	}

	wg.Wait()
}

// func TestFit(t *testing.T) {
// 	conf := testConfig()
// 	clientset, _ := kubernetes.GetClient(false, conf.Kubeconfig)
// 	taskCondition := &ResourceQuotaManager{clientset, conf}
// 	id, _ := uuid.NewUUID()
// 	task := model.Task{
// 		ID:          id.String(),
// 		Name:        "test-task",
// 		DriverCPU:   1,
// 		DriverMem:   500,
// 		ExecutorCPU: 1,
// 		ExecutorMem: 500,
// 		ExecutorNum: 1,
// 		Class:       "org.apache.spark.examples.SparkPi",
// 		TaskArgs:    []string{"http://172.23.4.148/oam/jars/spark-examples_2.11-2.3.0.jar", "10"},
// 	}
// 	if !taskCondition.Fit(&task) {
// 		t.Errorf("lack of resource")
// 	}
// }
