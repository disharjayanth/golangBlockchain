package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func GetBlocksAfter(blockHash Hash, dataDir string) ([]Block, error) {
	f, err := os.OpenFile(getBlockDBFilePath(dataDir), os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("error while opening blocks.db file in GetBlocksAfer func: %w", err)
	}

	blocks := make([]Block, 0)

	shouldStartCollecting := false

	if reflect.DeepEqual(blockHash, Hash{}) {
		shouldStartCollecting = true
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error while scanning block.db file in GetBlocksAfter func: %w", err)
		}

		var blocksFS BlockFS
		err = json.Unmarshal(scanner.Bytes(), &blocksFS)
		if err != nil {
			return nil, fmt.Errorf("error while unmarshalling to BlockFS in GetBlockAfter func: %w", err)
		}

		if shouldStartCollecting {
			blocks = append(blocks, blocksFS.Value)
		}

		if blockHash == blocksFS.Key {
			shouldStartCollecting = true
		}
	}

	return blocks, nil
}
