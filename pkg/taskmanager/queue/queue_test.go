package queue

import (
	"fmt"
	"testing"

	"github.com/TalkingData/hummingbird/pkg/taskmanager/model"
	v3 "go.etcd.io/etcd/clientv3"
)

func TestTaskQueue(t *testing.T) {
	config := v3.Config{Endpoints: []string{"localhost:2379"}}
	q, err := NewQueue(config, "testqueue")
	if err != nil {
		t.Error(err)
		return
	}

	num := 3
	for i := 1; i <= num; i++ {
		testTask := model.Task{Name: fmt.Sprintf("test-task-%d", i)}
		if err := q.Enqueue(&testTask); err != nil {
			t.Error(err)
			return
		}
	}

	for i := 1; i <= num; i++ {
		testTask, err := q.Dequeue()
		if err != nil {
			t.Error(err)
			return
		}
		expectedTaskName := fmt.Sprintf("test-task-%d", i)
		if testTask.Name != expectedTaskName {
			t.Errorf("Expected the task %s but instead got %s", expectedTaskName, testTask.Name)
			return
		}
	}
}
