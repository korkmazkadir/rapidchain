package network

import (
	"fmt"
	"net/rpc"

	"github.com/korkmazkadir/rapidchain/common"
)

// Client implements P2P client
type P2PClient struct {
	IPAddress  string
	portNumber int
	nodeID     int

	rpcClient *rpc.Client

	blockChunks    chan common.BlockChunk
	votes          chan common.Vote
	stateUpdates   chan common.StateUpdate
	connectRequest chan common.ConnectRequest

	stateUpdateMap map[string]struct{}

	err error
}

// NewClient creates a new client
func NewClient(IPAddress string, portNumber int, nodeID int) (*P2PClient, error) {

	rpcClient, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", IPAddress, portNumber))
	if err != nil {
		return nil, err
	}

	client := &P2PClient{}
	client.IPAddress = IPAddress
	client.portNumber = portNumber
	client.nodeID = nodeID
	client.rpcClient = rpcClient

	client.blockChunks = make(chan common.BlockChunk, 1024)
	client.votes = make(chan common.Vote, 1024)
	client.stateUpdates = make(chan common.StateUpdate, 1024)
	client.connectRequest = make(chan common.ConnectRequest, 1)

	client.stateUpdateMap = make(map[string]struct{})

	return client, nil
}

// Start starts the main loop of client. It blocks the calling goroutine
func (c *P2PClient) Start() {

	c.mainLoop()
}

// SendBlockChunk enques a chunk of a block to send
func (c *P2PClient) SendBlockChunk(chunk common.BlockChunk) {

	c.blockChunks <- chunk
}

// SendVote enques a vote to send
func (c *P2PClient) SendVote(vote common.Vote) {

	c.votes <- vote
}

func (c *P2PClient) SendStateUpdate(update common.StateUpdate) {

	c.stateUpdates <- update
}

func (c *P2PClient) SendConnecRequest(connectRequest common.ConnectRequest) {

	c.connectRequest <- connectRequest
}

func (c *P2PClient) mainLoop() {

	for {
		select {

		case connectRequest := <-c.connectRequest:
			go c.rpcClient.Call("P2PServer.HandleConnectRequest", connectRequest, nil)

		case stateUpdate := <-c.stateUpdates:
			go c.rpcClient.Call("P2PServer.HandleStateUpdate", stateUpdate, nil)

		case vote := <-c.votes:
			go c.rpcClient.Call("P2PServer.HandleVote", vote, nil)

		case blockChunk := <-c.blockChunks:
			go c.rpcClient.Call("P2PServer.HandleBlockChunk", blockChunk, nil)

		}
	}
}
