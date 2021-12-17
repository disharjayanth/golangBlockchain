package main

import (
	"fmt"
	"os"

	"github.com/disharjayanth/golangBlockchain/database"
	"github.com/spf13/cobra"
)

var balancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all balances",
	Run: func(cmd *cobra.Command, args []string) {
		state, err := database.NewStateFromDisk()
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(1)
		}
		defer state.Close()

		fmt.Println("Account Balances")
		fmt.Println("----------------")
		fmt.Println("")

		for account, balanace := range state.Balances {
			fmt.Printf("%s: %d\n", account, balanace)
		}
	},
}

func balancesCmd() *cobra.Command {
	var balancesCmd = &cobra.Command{
		Use:   "balances",
		Short: "Interact with balances (list)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	balancesCmd.AddCommand(balancesListCmd)

	return balancesCmd
}
