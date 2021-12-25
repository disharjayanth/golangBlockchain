package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/disharjayanth/golangBlockchain/database"
)

const httpPort = 8000

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

func Run(dataDir string) error {
	fmt.Printf("Listening on HTTP port: %d\n", httpPort)

	state, err := database.NewStateFromDisk(dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)
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

func writeRes(w http.ResponseWriter, content interface{}) {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		writeErrRes(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contentJSON)
}

func readReq(r *http.Request, reqBody interface{}) error {
	reqBodyJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error while reading request JSON body in readReq func: %w", err)
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJSON, reqBody)
	if err != nil {
		return fmt.Errorf("error whuke unmarshalling req body in readReq func: %w", err)
	}

	return nil
}

func writeErrRes(w http.ResponseWriter, err error) {
	jsonErrRes, _ := json.Marshal(ErrRes{err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonErrRes)
}
