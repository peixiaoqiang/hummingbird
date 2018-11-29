package taskmanager

import (
	"encoding/json"
	"testing"

	"github.com/TalkingData/hummingbird/pkg/kubernetes"
	"github.com/TalkingData/hummingbird/pkg/taskmanager/model"
	"github.com/google/uuid"
	check "gopkg.in/check.v1"
	k8s "k8s.io/client-go/kubernetes"
)

func Test(t *testing.T) { check.TestingT(t) }

type TaskManagerTestSuit struct {
}

var _ = check.Suite(&TaskManagerTestSuit{})

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

func newFakeTaskManger(client *k8s.Clientset, conf *Conf) (*TaskManagerImpl, map[string]*FakeTaskQueue, error) {
	qmap := map[string]*FakeTaskQueue{}
	waiting := newFakeQueue()
	running := newFakeQueue()
	failed := newFakeQueue()
	finish := newFakeQueue()
	qmap["waiting"] = waiting
	qmap["running"] = running
	qmap["failed"] = failed
	qmap["finish"] = finish

	resourceQuotaManager, _ := NewFakeResourceQuotaManager(conf)
	quota := &ResourceQuota{LimitCPU: 2, LimitMem: 1000}
	resourceQuotaManager.Init(conf.ResourceQuotaName, quota)

	return &TaskManagerImpl{resourcePool: resourceQuotaManager, waiting: waiting, running: running, failed: failed, finish: finish, taskEngine: &FakeTaskEngine{}}, qmap, nil
}

func newFakeQueue() *FakeTaskQueue {
	return &FakeTaskQueue{queue: make([]*model.Task, 0), index: map[string]*model.Task{}}
}

type FakeTaskEngine struct {
}

func (f *FakeTaskEngine) Submit(model.Task) error {
	return nil
}

func (f *FakeTaskEngine) GetTaskState(*model.Task) string {
	return ""
}

type FakeTaskQueue struct {
	queue []*model.Task
	index map[string]*model.Task
}

func (q *FakeTaskQueue) Enqueue(m *model.Task) error {
	q.queue = append(q.queue, m)
	q.index[m.Name] = m
	return nil
}

func (q *FakeTaskQueue) Dequeue() (*model.Task, error) {
	if len(q.queue) > 0 {
		element := q.queue[0]
		delete(q.index, element.Name)

		if len(q.queue) == 1 {
			q.queue = q.queue[1:]
		} else {
			q.queue = make([]*model.Task, 0)
		}
		return element, nil
	}
	return nil, nil
}

func (q *FakeTaskQueue) Get(name string) (*model.Task, error) {
	return q.index[name], nil
}

func (q *FakeTaskQueue) GetFirst() (*model.Task, error) {
	if len(q.queue) > 0 {
		element := q.queue[0]
		return element, nil
	}
	return nil, nil
}

func (q *FakeTaskQueue) Remove(name string) error {
	for i, m := range q.queue {
		if m.Name == name {
			delete(q.index, name)
			if len(q.queue) > 1 {
				q.queue = append(q.queue[0:i], q.queue[i+1:]...)
			} else {
				q.queue = make([]*model.Task, 0)
			}
			break
		}
	}
	return nil
}

func (t *TaskManagerTestSuit) TestAddTask(c *check.C) {
	conf := testConfig()
	clientset, _ := kubernetes.GetClient(false, conf.Kubeconfig)
	tm, qmap, _ := newFakeTaskManger(clientset, conf)
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
	err := tm.Run(task, nil)
	c.Assert(err, check.IsNil)

	expected, err := qmap["running"].GetFirst()
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.NotNil)
	c.Assert(task.ID, check.Equals, expected.ID)
}

func (t *TaskManagerTestSuit) TestAddTaskNoResQuota(c *check.C) {
	conf := testConfig()
	clientset, _ := kubernetes.GetClient(false, conf.Kubeconfig)
	tm, qmap, _ := newFakeTaskManger(clientset, conf)
	id, _ := uuid.NewUUID()
	task := model.Task{
		ID:          id.String(),
		Name:        "test-task",
		DriverCPU:   4,
		DriverMem:   2000,
		ExecutorCPU: 4,
		ExecutorMem: 2000,
		ExecutorNum: 4,
		Class:       "org.apache.spark.examples.SparkPi",
		TaskArgs:    []string{"http://172.23.4.148/oam/jars/spark-examples_2.11-2.3.0.jar", "10"},
	}
	err := tm.Run(task, nil)
	c.Assert(err, check.IsNil)

	expected, err := qmap["waiting"].GetFirst()
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.NotNil)
	c.Assert(task.ID, check.Equals, expected.ID)
}

func (t *TaskManagerTestSuit) TestPickTask(c *check.C) {
	conf := testConfig()
	clientset, _ := kubernetes.GetClient(false, conf.Kubeconfig)
	tm, qmap, _ := newFakeTaskManger(clientset, conf)
	id, _ := uuid.NewUUID()
	task := model.Task{
		ID:          id.String(),
		Name:        "test-task",
		DriverCPU:   1,
		DriverMem:   1,
		ExecutorCPU: 1,
		ExecutorMem: 1,
		ExecutorNum: 1,
		Class:       "org.apache.spark.examples.SparkPi",
		TaskArgs:    []string{"http://172.23.4.148/oam/jars/spark-examples_2.11-2.3.0.jar", "10"},
	}

	qmap["waiting"].Enqueue(&task)
	err := tm.Pick(nil)
	c.Assert(err, check.IsNil)

	expected, err := qmap["running"].GetFirst()
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.NotNil)
	c.Assert(task.ID, check.Equals, expected.ID)
}

// func TestAddTask(t *testing.T) {
// 	conf := testConfig()
// 	clientset, err := kubernetes.GetClient(false, conf.Kubeconfig)
// 	taskManager, _ := NewTaskManager(clientset, conf)

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

// 	finish := make(chan struct{}, 1)
// 	defer close(finish)
// 	err = taskManager.Run(task, finish)
// 	if err != nil {
// 		t.Errorf("fail to start task: %v", err)
// 		return
// 	}
// 	<-finish
// }

// func TestPick(t *testing.T) {
// 	conf := testConfig()
// 	clientset, _ := kubernetes.GetClient(false, conf.Kubeconfig)
// 	taskManager, _ := NewTaskManager(clientset, conf)
// 	finish := make(chan struct{})
// 	err := taskManager.Pick(finish)
// 	if err != nil {
// 		t.Errorf("fail to pick task: %v", err)
// 		return
// 	}
// 	<-finish
// }

// func TestBatchAddTask(t *testing.T) {
// 	conf := testConfig()
// 	clientset, err := kubernetes.GetClient(false, conf.Kubeconfig)
// 	taskManager, _ := NewTaskManager(clientset, conf)

// 	num := 10
// 	wg := sync.WaitGroup{}
// 	wg.Add(num)
// 	for i := 0; i < num; i++ {
// 		id, _ := uuid.NewUUID()
// 		task := model.Task{
// 			ID:          id.String(),
// 			Name:        "test-task",
// 			DriverCPU:   1,
// 			DriverMem:   500,
// 			ExecutorCPU: 1,
// 			ExecutorMem: 500,
// 			ExecutorNum: 1,
// 			Class:       "org.apache.spark.examples.SparkPi",
// 			TaskArgs:    []string{"http://172.23.4.148/oam/jars/spark-examples_2.11-2.3.0.jar", "10"},
// 		}

// 		go func() {
// 			defer wg.Done()
// 			finish := make(chan struct{}, 1)
// 			defer close(finish)
// 			err = taskManager.Run(task, finish)
// 			if err != nil {
// 				t.Errorf("fail to start task: %v", err)
// 			}
// 			<-finish
// 		}()
// 	}

// 	wg.Wait()
// }

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
