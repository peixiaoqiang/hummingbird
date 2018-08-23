# Spark Controller
Spark Controller is a controller based on Kubernetes Informer that enable to watch spark driver pod to retrieve and store spark application information. There is also a http server to access.

## Getting Started
Run controller and http server:

```
go run spark/starter.go -conf config.conf
```

Or you can build docker image and then start it in Kubernetes using yaml:

```
$ bash tools/spark-controller/build.sh spark-controller:v0.1 <repo-server>
// Please change the namespace you want to watch
$ kubectl apply -f tools/spark-controller.yaml
```

You are able to get spark application information by:

```
$ curl http://localhost:9001/applications/<spark-driver-pod-name>
{
  "id": "spark-application-1533625518534",
  "name": "spark-pi-6d4aefa29db238bf814339901c6499c9-driver",
  "jobs": [
    {
      "name": "reduce at SparkPi.scala:38",
      "status": "RUNNING",
      "numActiveTasks": 2,
      "numTasks": 500
    }
  ],
  "attempts": [
    {
      "startTime": "2018-08-07T07:05:16.654GMT",
      "endTime": "1969-12-31T23:59:59.999GMT"
    }
  ]
}
```
## How it works
Spark Controller starts two goroutine, they are **spark controller** and **http server**.
### Spark Controller
Spark Controller use Kubernetes Informer to observe Kubernetes pods events including `Add` and `Update`:

```
github.com/TalkingData/hummingbird/pkg/spark/controller.go:18

func Run(clientset *kubernetes.Clientset, namespace string, podCB PodCallback, stopCh <-chan struct{}) {
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), string(v1.ResourcePods), namespace,
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
			   // Handle pod add 
				...
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				// Handle pod update
				...
			},
		},
	)

	controller.Run(stopCh)
}
```
When one runs a spark application on Kuberentes, `AddFunc` will be triggered. `Watch` will filter spark driver pod out and wait it to become running. Then, it will callback interface which impelements `PodCallback`:

```
github.com/TalkingData/hummingbird/pkg/spark/controller.go:13

type PodCallback interface {
	OnAddRunningPod(pod *v1.Pod)
	OnUpdatePod(oldPod *v1.Pod, newPod *v1.Pod)
}
```
The default is `ApplicationHandler` in `pkg/spark/spark.go`. This handler is designed to retrieve spark application information from spark driver ui, which will start in spark driver pod on `<pod_ip>:4040`. The most important attribute is **spark application id**. At last, the handler stores the information to the storage which is etcd by default. 

The process of handling `Update` event is simple, it just store the status of spark driver pod.

### Http Server
Controller Server serves http api request, it only has one handler `handleApplication` routing `/applications/<spark-driver-pod-name>`. When receives request, the function will first get spark application id from the storage by `spark-driver-pod-name`. Subsequently, if the driver pod is running, it will get spark application information from spark driver ui as same as above. Otherwise, the information will be retrieved from spark history which is setup in advance.

## License
The project is Open Source software released under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0.html).


