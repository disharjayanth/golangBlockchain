package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var genesisJSON = `{
	"genesis_time": "2019-03-18T00:00:00.000000000Z",
	"chain_id": "the-blockchain-bar-ledger",
	"balances": {
	  "andrej": 1000000
	}
}`

type Genesis struct {
	Balances map[Account]uint `json:"balances"`
}

func LoadGenesis(path string) (Genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Genesis{}, fmt.Errorf("error while reading genesisDb file: %w", err)
	}

	var loadedGenesis Genesis
	if err = json.Unmarshal(content, &loadedGenesis); err != nil {
		return Genesis{}, fmt.Errorf("error while unmarshalling genesisblock to struct: %w", err)
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(genesisJSON), 0644)
}
