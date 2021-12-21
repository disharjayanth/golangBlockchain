package main

import (
	"fmt"
	"os"
	"time"

	"github.com/disharjayanth/golangBlockchain/database"
)

func main() {
	cwd, _ := os.Getwd()
	state, err := database.NewStateFromDisk(cwd)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(database.Hash{}, uint64(time.Now().Unix()), []database.Tx{
		database.NewTx("andrej", "andrej", 3, ""),
		database.NewTx("andrej", "andrej", 700, "reward"),
	})

	state.AddBlock(block0)

	block0Hash, _ := state.Persist()

	fmt.Println("block0Hash", block0Hash)

	block1 := database.NewBlock(block0Hash, uint64(time.Now().Unix()), []database.Tx{
		database.NewTx("andrej", "babayaga", 2000, ""),
		database.NewTx("andrej", "andrej", 100, "reward"),
		database.NewTx("babayaga", "andrej", 1, ""),
		database.NewTx("babayaga", "caesar", 1000, ""),
		database.NewTx("babayaga", "andrej", 50, ""),
		database.NewTx("andrej", "andrej", 600, "reward"),
	})

	state.AddBlock(block1)

	block1Hash, _ := state.Persist()

	fmt.Println("block1Hash", block1Hash)
}
