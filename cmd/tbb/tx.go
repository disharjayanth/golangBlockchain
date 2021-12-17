package main

import (
	"fmt"
	"os"

	"github.com/disharjayanth/golangBlockchain/database"
	"github.com/spf13/cobra"
)

const flagFrom = "from"
const flagTo = "to"
const flagValue = "value"
const flagData = "data"

func txCmd() *cobra.Command {
	var txCmd = &cobra.Command{
		Use:   "tx",
		Short: "Interact with txs (add....).",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	txCmd.AddCommand(txAddCmd())

	return txCmd
}

func txAddCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add",
		Short: "Add new Tx to database",
		Run: func(cmd *cobra.Command, args []string) {
			from, _ := cmd.Flags().GetString(flagFrom)
			to, _ := cmd.Flags().GetString(flagTo)
			value, _ := cmd.Flags().GetUint(flagValue)
			data, _ := cmd.Flags().GetString(flagData)

			fromAcc := database.NewAccount(from)
			toAcc := database.NewAccount(to)

			tx := database.NewTx(fromAcc, toAcc, value, data)

			state, err := database.NewStateFromDisk()
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				os.Exit(1)
			}
			defer state.Close()

			// Add tx to in memory slice (pool)
			err = state.Add(tx)
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				os.Exit(1)
			}

			// Flush the mempool tx to disk
			err = state.Persist()
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				os.Exit(1)
			}

			fmt.Println("Tx successfully added to the ledger!")
		},
	}

	cmd.Flags().String(flagFrom, "", "From what account to send tokens")
	cmd.MarkFlagRequired(flagFrom)

	cmd.Flags().String(flagTo, "", "To what account to send token")
	cmd.MarkFlagRequired(flagTo)

	cmd.Flags().Uint(flagValue, 0, "How many tokens to send")
	cmd.MarkFlagRequired(flagValue)

	cmd.Flags().String(flagData, "", `Possible values: "reward"`)

	return cmd
}
