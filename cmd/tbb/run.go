package main

import (
	"fmt"
	"os"

	"github.com/disharjayanth/golangBlockchain/node"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches TBB node and its HTTP API.",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, _ := cmd.Flags().GetString(flagDataDir)

			fmt.Println("Launching TBB node and its HTTP API....")

			bootstrap := node.NewPeerNode(getDataDirFromCmd(cmd), 8001, true, true)

			n := node.New(dataDir, 8000, bootstrap)

			err := n.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(runCmd)
	runCmd.Flags().Uint64(flagPort, node.DefaultHTTPPort, "exposed HTTP port for communication with peers")

	return runCmd
}
