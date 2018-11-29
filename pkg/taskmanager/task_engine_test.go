package taskmanager

// import (
// 	"testing"

// 	"github.com/TalkingData/hummingbird/pkg/taskmanager/model"
// 	"github.com/google/uuid"
// )

// func TestSparkOnK8SEngine(t *testing.T) {
// 	s := SparkOnK8SEngine{DeployMode: "cluster", MasterIP: "k8s://https://172.20.0.67:6443", Image: "172.20.0.112/library/spark_on_k8s:v1.5", NameSpace: "test-spark", ServiceAccoutName: "test-spark"}
// 	id, err := uuid.NewUUID()
// 	task := model.Task{
// 		ID:          id.String(),
// 		Name:        "test-task",
// 		DriverCPU:   1,
// 		DriverMem:   1000,
// 		ExecutorCPU: 1,
// 		ExecutorMem: 1000,
// 		ExecutorNum: 1,
// 		Class:       "org.apache.spark.examples.SparkPi",
// 		TaskArgs:    []string{"http://172.23.4.148/oam/jars/spark-examples_2.11-2.3.0.jar", "10"},
// 	}
// 	err = s.Submit(task)
// 	if err != nil {
// 		t.Errorf("%v", err)
// 	}
// }
