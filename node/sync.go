package node

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func (n *Node) sync(ctx context.Context) error {
	ticker := time.NewTicker(45 * time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Println("Searching for new Peers and Blocks....")
			n.fetchNewBlocksAndPeers()

		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) doSync() {
	for _, peer := range n.knownPeers {
		status, err := queryPeerStatus(peer)
		err = n.joinKnownPeers(peer)
		err = n.syncBlocks(peer)
		err = n.syncKnownPeers(peer)
	}
}

func (n *Node) fetchNewBlocksAndPeers() {
	for _, peer := range n.knownPeers {
		status, err := queryPeerStatus(peer)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue
		}

		localBlockNumber := n.state.LatestBlock().Header.Number

		if localBlockNumber < status.Number {
			newBlocksNumber := status.Number - localBlockNumber
			fmt.Printf("Found %d new blocks from Peer %s\n", newBlocksNumber, peer.IP)
		}

		for _, statusPeer := range status.KnownPeers {
			_, isKnownPeer := n.knownPeers[statusPeer.TcpAddress()]
			if !isKnownPeer {
				fmt.Printf("Found new Peer %s\n", peer.TcpAddress())

				n.knownPeers[statusPeer.TcpAddress()] = statusPeer
			}
		}
	}
}

func queryPeerStatus(peer PeerNode) (StatusRes, error) {
	url := fmt.Sprintf("http://%s/%s", peer.TcpAddress(), endPointStatus)
	res, err := http.Get(url)
	if err != nil {
		return StatusRes{}, fmt.Errorf("error while checking peer status in queryPeerStatus func: %w", err)
	}

	statusRes := StatusRes{}
	err = readRes(res, statusRes)
	if err != nil {
		return StatusRes{}, err
	}

	return statusRes, nil
}
