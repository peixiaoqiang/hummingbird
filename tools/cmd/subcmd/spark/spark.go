package spark

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TalkingData/hummingbird/tools/cmd/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	Spark = &cobra.Command{
		Use:   "spark",
		Short: "",
		Run:   run,
	}
	kubeconf  string
	action    string
	namespace string
)

func init() {
	Spark.Flags().StringVar(&kubeconf, "kubeconf", "", "")
	Spark.Flags().StringVar(&action, "action", "", "")
	Spark.Flags().StringVar(&namespace, "namespace", "", "")
}

func run(cmd *cobra.Command, args []string) {
	switch action {
	case "tasks":
		listTasks()
	default:
		fmt.Println("wrong action")
		return
	}
}

func listTasks() {
	client, err := util.GetKubeClient(kubeconf)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	data := [][]string{}
	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "spark-role=driver"})
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	now := time.Now()
	for _, p := range pods.Items {
		duration := now.Sub(p.Status.StartTime.Time)
		taskName := strings.Replace(p.Name, "-driver", "", 1)
		data = append(data, []string{namespace, taskName, formatDuration(duration), string(p.Status.Phase)})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "POD NAME", "DURATION", "STATUS"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

func formatDuration(d time.Duration) string {
	if d < day {
		return d.String()
	}

	var b strings.Builder
	if d >= year {
		years := d / year
		fmt.Fprintf(&b, "%dy", years)
		d -= years * year
	}

	days := d / day
	d -= days * day
	fmt.Fprintf(&b, "%dd%d", days, int(d.Seconds()))

	return b.String()
}
