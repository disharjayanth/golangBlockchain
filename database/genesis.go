package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

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
