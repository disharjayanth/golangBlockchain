package main

import (
	"fmt"
	"os"
	"time"

	"github.com/disharjayanth/golangBlockchain/database"
	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	var migateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain database according to new business rules",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				os.Exit(1)
			}
			defer state.Close()

			block0 := database.NewBlock(database.Hash{}, state.NextBlockNumber(), uint64(time.Now().Unix()), []database.Tx{
				database.NewTx("andrej", "andrej", 3, ""),
				database.NewTx("andrej", "andrej", 700, "reward"),
			})

			block0Hash, err := state.AddBlock(block0)
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				os.Exit(1)
			}

			block1 := database.NewBlock(block0Hash, state.NextBlockNumber(), uint64(time.Now().Unix()), []database.Tx{
				database.NewTx("andrej", "babayaga", 2000, ""),
				database.NewTx("andrej", "andrej", 100, "reward"),
				database.NewTx("babayaga", "andrej", 1, ""),
				database.NewTx("babayaga", "caesar", 1000, ""),
				database.NewTx("babayaga", "andrej", 50, ""),
				database.NewTx("andrej", "andrej", 600, "reward"),
			})

			block1Hash, err := state.AddBlock(block1)
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				os.Exit(1)
			}

			block2 := database.NewBlock(block1Hash, state.NextBlockNumber(), uint64(time.Now().Unix()), []database.Tx{
				database.NewTx("andrej", "andrej", 24700, "reward"),
			})

			_, err = state.AddBlock(block2)
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migateCmd)

	return migateCmd
}
