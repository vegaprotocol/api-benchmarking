package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vegaprotocol/datanode-api-benchmarking/cmd/orders"
)

var rootCmd = &cobra.Command{
	Use:   "vega-api-bench",
	Short: "A benchmarking tool for vega APIs",
	Long:  "A tool for benchmarking vega APIs",
	Run:   run,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(orders.ListOrdersCmd)
}

func run(cmd *cobra.Command, args []string) {
}
