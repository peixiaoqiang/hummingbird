package main

import (
	"fmt"
	"os"

	"github.com/TalkingData/hummingbird/tools/cmd/subcmd/usage"
	"github.com/TalkingData/hummingbird/tools/cmd/subcmd/yarn"
	"github.com/TalkingData/hummingbird/tools/cmd/subcmd/spark"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	rootCmd.AddCommand(usage.Usage)
	rootCmd.AddCommand(yarn.Yarn)
	rootCmd.AddCommand(spark.Spark)
	Execute()
}
