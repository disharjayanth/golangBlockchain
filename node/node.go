package node

import (
	"fmt"
	"net/http"

	"github.com/disharjayanth/golangBlockchain/database"
)

const DefaultHTTPPort = 8000

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`

	// Whenever my node already established connection, sync with this peer
	connected bool
}

type Node struct {
	dataDir string
	port    uint64

	// To inject state into HTTP handler
	state *database.State

	knownPeers []PeerNode
}

func New(dataDir string, port uint64, bootstrap PeerNode) *Node {
	return &Node{
		dataDir:    dataDir,
		port:       port,
		knownPeers: []PeerNode{bootstrap},
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, connected bool) PeerNode {
	return PeerNode{
		IP:          ip,
		Port:        port,
		IsBootstrap: isBootstrap,
		connected:   connected,
	}
}

func (n *Node) Run() error {
	fmt.Println("Listening on HTTP port: ", DefaultHTTPPort)

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	n.state = state

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	http.HandleFunc("/node/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}
