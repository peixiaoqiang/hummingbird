package yarn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/TalkingData/hummingbird/tools/cmd/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	Yarn = &cobra.Command{
		Use:   "yarn",
		Short: "",
		Run:   run,
	}
	kubeconf  string
	confPath  string
	action    string
	namespace string
)

func init() {
	Yarn.Flags().StringVar(&kubeconf, "kubeconf", "", "")
	Yarn.Flags().StringVar(&confPath, "conf", "", "")
	Yarn.Flags().StringVar(&action, "action", "", "")
	Yarn.Flags().StringVar(&namespace, "namespace", "", "")
}

type Conf struct {
	Namespaces []string `json:"namespaces"`
}

var CONF *Conf

func readConf() (err error) {
	CONF = &Conf{}
	data, err := ioutil.ReadFile(confPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, CONF)
	return
}

func run(cmd *cobra.Command, args []string) {
	err := readConf()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	switch action {
	case "activePods":
		activePods()
	case "podDetail":
		podDetail(namespace, args[0])
	default:
		fmt.Println("wrong action")
		return
	}
}

func activePods() {
	client, err := util.GetKubeClient(kubeconf)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	data := [][]string{}
	for _, ns := range CONF.Namespaces {
		pods, err := client.CoreV1().Pods(ns).List(metav1.ListOptions{FieldSelector: "status.phase=Running"})
		if err != nil {
			fmt.Printf("%v", err)
			return
		}
		data = append(data, []string{ns, fmt.Sprintf("%v", len(pods.Items))})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "ACTIVE_PODS"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func podDetail(ns string, podName string) {
	client, err := util.GetKubeClient(kubeconf)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	pod, err := client.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	data := [][]string{[]string{ns, podName, pod.Status.PodIP, pod.Spec.NodeName}}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "NAME", "IP", "HOST"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
