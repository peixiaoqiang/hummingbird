package usage

import (
	"fmt"
	"os"

	kube "github.com/TalkingData/hummingbird/pkg/kubernetes"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	Usage = &cobra.Command{
		Use:   "usage",
		Short: "",
		Run:   run,
	}
	kubeconf  string
	namespace string
	t         string
)

func init() {
	Usage.Flags().StringVar(&kubeconf, "kubeconf", "", "")
	Usage.Flags().StringVar(&namespace, "namespace", "", "")
	Usage.Flags().StringVar(&t, "type", "ns", "")
}

type podItem struct {
	resourceItem
	id   string
	name string
}

type resourceItem struct {
	requestCPU resource.Quantity
	limitCPU   resource.Quantity
	requestMem resource.Quantity
	limitMem   resource.Quantity
}

func run(cmd *cobra.Command, args []string) {
	_, sumNS, err := doRun(namespace, kubeconf)
	if err != nil {
		fmt.Printf("%v", err)
	}

	switch t {

	case "pod":
	case "ns":
		renderNS(sumNS)
	}
}

func doRun(ns string, kubeConf string) (sumPods map[string][]podItem, sumNS map[string]resourceItem, err error) {
	sumPods = map[string][]podItem{}
	sumNS = map[string]resourceItem{}
	var client *kubernetes.Clientset
	if kubeConf != "" {
		client, err = kube.GetClient(false, kubeConf)
	} else {
		client, err = kube.GetClient(true, "")
	}
	if err != nil {
		return
	}

	if ns == "" {
		namespaces, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for _, nsItem := range namespaces.Items {
			sum, err := convert(client, nsItem.Name)
			if err != nil {
				return nil, nil, err
			}
			sumPods[nsItem.Name] = sum
		}
	} else {
		sum, err := convert(client, ns)
		if err != nil {
			return nil, nil, err
		}
		fmt.Printf("%v", ns)
		sumPods[ns] = sum
	}

	requestCPU, requestMem, limitCPU, limitMem := resource.Quantity{}, resource.Quantity{}, resource.Quantity{}, resource.Quantity{}
	for k, v := range sumPods {
		for _, pItem := range v {
			requestCPU.Add(pItem.requestCPU)
			requestMem.Add(pItem.requestMem)
			limitCPU.Add(pItem.limitCPU)
			limitMem.Add(pItem.limitMem)
		}
		sumNS[k] = resourceItem{requestCPU: requestCPU, requestMem: requestMem, limitCPU: limitCPU, limitMem: limitMem}
	}

	return sumPods, sumNS, nil
}

func convert(client *kubernetes.Clientset, ns string) (podItems []podItem, err error) {
	pods, err := client.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		return
	}

	for _, p := range pods.Items {
		requestCPU, requestMem, limitCPU, limitMem := resource.Quantity{}, resource.Quantity{}, resource.Quantity{}, resource.Quantity{}
		for _, c := range p.Spec.Containers {
			requestCPU.Add(*c.Resources.Requests.Cpu())
			requestMem.Add(*c.Resources.Requests.Memory())
			limitCPU.Add(*c.Resources.Limits.Cpu())
			limitMem.Add(*c.Resources.Limits.Memory())
		}
		pItem := podItem{name: p.Name, id: string(p.UID)}
		pItem.requestCPU = requestCPU
		pItem.requestMem = requestMem
		pItem.limitCPU = limitCPU
		pItem.limitMem = limitMem
		podItems = append(podItems, pItem)
	}
	return
}

func renderNS(sumNS map[string]resourceItem) {
	data := [][]string{}
	for k, v := range sumNS {
		data = append(data, []string{k, v.requestCPU.String(), v.limitCPU.String(), v.requestMem.String(), v.limitMem.String()})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "REQUEST_CPU", "LIMIT_CPU", "REQUEST_MEM", "LIMIT_MEM"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
