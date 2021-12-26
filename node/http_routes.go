package node

import (
	"net/http"
	"time"

	"github.com/disharjayanth/golangBlockchain/database"
)

type ErrRes struct {
	Error string `json:"error"`
}

type BalancesRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type TxAddReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddRes struct {
	Hash database.Hash `json:"block_hash"`
}

type StatusRes struct {
	Hash       database.Hash `json:"block_hash"`
	Number     uint64        `json:"block_number"`
	KnownPeers []PeerNode    `json:"peer_known"`
}

func listBalancesHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeRes(w, BalancesRes{state.LatestBlockHash(), state.Balances})
}

func txAddHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	req := TxAddReq{}
	err := readReq(r, &req)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	tx := database.NewTx(database.NewAccount(req.From), database.NewAccount(req.To), req.Value, req.Data)

	block := database.NewBlock(state.LatestBlockHash(), state.NextBlockNumber(), uint64(time.Now().Unix()), []database.Tx{tx})

	hash, err := state.AddBlock(block)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, TxAddRes{Hash: hash})
}

func statusHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	res := StatusRes{
		Hash:       node.state.LatestBlockHash(),
		Number:     node.state.LatestBlock().Header.Number,
		KnownPeers: node.knownPeers,
	}

	writeRes(w, res)
}
