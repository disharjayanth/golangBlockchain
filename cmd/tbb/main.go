package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const flagDataDir = "datadir"

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path to the node dat dir where DB will/is stored")
	cmd.MarkFlagRequired(flagDataDir)
}

func main() {
	tbbCmd := &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	tbbCmd.AddCommand(versionCmd)
	tbbCmd.AddCommand(runCmd())
	tbbCmd.AddCommand(balancesCmd())
	tbbCmd.AddCommand(txCmd())

	if err := tbbCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
