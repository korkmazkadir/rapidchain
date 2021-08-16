package network

import (
	"errors"
	"log"
	"sync"

	"github.com/korkmazkadir/rapidchain/common"
)

var NoCorrectPeerAvailable = errors.New("there are no correct peers available")

type PeerSet struct {
	outPeers   []*P2PClient
	outPeerMap map[int]*P2PClient
	demux      *common.Demux

	ipAddress  string
	portNumber int
	nodeID     int

	savedCount int

	// protects following fields
	mutex   sync.Mutex
	inPeers []*P2PClient
}

func NewPeerSet(demux *common.Demux, ipAddress string, portNumber int, currentNodeID int) *PeerSet {
	peerSet := &PeerSet{
		demux:      demux,
		ipAddress:  ipAddress,
		portNumber: portNumber,
		nodeID:     currentNodeID,
		outPeerMap: make(map[int]*P2PClient),
	}

	// runs a go routine to handle connection requests from in peers
	go peerSet.handleConnectRequests()

	return peerSet
}

func (p *PeerSet) AddPeer(IPAddress string, portNumber int, nodeID int) error {

	client, err := NewClient(IPAddress, portNumber, nodeID)
	if err != nil {
		return err
	}

	// starts the main loop of client
	go client.Start()

	p.outPeers = append(p.outPeers, client)
	p.outPeerMap[nodeID] = client
	// sending connect request here
	client.SendConnecRequest(common.ConnectRequest{NodeID: p.nodeID, IPAddress: p.ipAddress, PortNumber: p.portNumber})

	return nil
}

func (p *PeerSet) DissaminateChunks(chunks []common.BlockChunk) {

	for index, chunk := range chunks {
		peer := p.selectPeer(index)
		peer.SendBlockChunk(chunk)
	}
}

func (p *PeerSet) ForwardChunk(chunk common.BlockChunk) {

	chunkHash := chunk.Hash()
	p.sendStateUpdate(chunkHash)
	p.receiveStateUpdates()

	forwardCount := 0
	for _, peer := range p.outPeers {
		if peer.err != nil {
			continue
		}
		forwardCount++

		if _, ok := peer.stateUpdateMap[string(chunkHash)]; ok {
			p.savedCount++
			continue
		}

		peer.SendBlockChunk(chunk)
	}

	if forwardCount == 0 {
		panic(NoCorrectPeerAvailable)
	}

}

func (p *PeerSet) ForwardVote(vote common.Vote) {

	forwardCount := 0
	for _, peer := range p.outPeers {
		if peer.err != nil {
			continue
		}
		forwardCount++
		peer.SendVote(vote)
	}

	if forwardCount == 0 {
		panic(NoCorrectPeerAvailable)
	}
}

func (p *PeerSet) GetSavedCount() int {
	return p.savedCount
}

func (p *PeerSet) ClearSavedCount() {
	p.savedCount = 0
}

func (p *PeerSet) selectPeer(index int) *P2PClient {

	peerCount := len(p.outPeers)
	for i := 0; i < peerCount; i++ {
		peer := p.outPeers[(index+i)%peerCount]
		if peer.err == nil {
			return peer
		}
	}

	panic(NoCorrectPeerAvailable)
}

/////////////////////////// in peers related

func (p *PeerSet) sendStateUpdate(messageHash []byte) {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, peer := range p.inPeers {
		if peer.err != nil {
			continue
		}

		peer.SendStateUpdate(common.StateUpdate{NodeID: p.nodeID, ReceivedMessageHash: messageHash})

	}

}

func (p *PeerSet) handleConnectRequests() {

	connectReqChan := p.demux.GetConnectRequestChan()

	for {
		// assumes each node sends a single connect message
		req := <-connectReqChan

		client, err := NewClient(req.IPAddress, req.PortNumber, req.NodeID)
		if err != nil {
			panic(err)
		}

		// starts the main loop of client
		go client.Start()

		p.mutex.Lock()
		p.inPeers = append(p.inPeers, client)
		p.mutex.Unlock()

		log.Printf("new in connection established %s:%d --> NodeID: %d\n", req.IPAddress, req.PortNumber, req.NodeID)

	}
}

func (p *PeerSet) receiveStateUpdates() {

	stateUpdateChan := p.demux.GetStateUpdateChan()

	for {
		select {
		case update := <-stateUpdateChan:

			//log.Printf("new state update from %d --> %s", update.NodeID, string(update.ReceivedMessageHash))

			outPeer, ok := p.outPeerMap[update.NodeID]
			if ok {
				// marks as received
				outPeer.stateUpdateMap[string(update.ReceivedMessageHash)] = struct{}{}
			}

		default:
			// chan emty so no need to wait
			return
		}
	}

}
