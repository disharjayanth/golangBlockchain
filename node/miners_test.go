package node

import (
	"encoding/hex"
	"testing"

	"github.com/disharjayanth/golangBlockchain/database"
)

func TestValidBlockHash(t *testing.T) {
	hexHash := "000000fa04f816039...a4db586086168edfa"
	var hash = database.Hash{}

	hex.Decode(hash[:], []byte(hexHash))

	isValid := database.IsBlockHashValid(hash)

	if !isValid {
		t.Fatalf("hash '%x' should be a valid hash", hexHash)
	}
}
