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

	rpcClient *rpc.Client

	blockChunks chan common.BlockChunk
	votes       chan common.Vote

	err error
}

// NewClient creates a new client
func NewClient(IPAddress string, portNumber int) (*P2PClient, error) {

	rpcClient, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", IPAddress, portNumber))
	if err != nil {
		return nil, err
	}

	client := &P2PClient{}
	client.IPAddress = IPAddress
	client.portNumber = portNumber
	client.rpcClient = rpcClient

	client.blockChunks = make(chan common.BlockChunk, 1024)
	client.votes = make(chan common.Vote, 1024)

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

func (c *P2PClient) mainLoop() {

	for {
		select {

		case vote := <-c.votes:
			c.rpcClient.Call("P2PServer.HandleVote", vote, nil)

		case blockChunk := <-c.blockChunks:
			c.rpcClient.Call("P2PServer.HandleBlockChunk", blockChunk, nil)

		}
	}
}
