package network

import (
	"fmt"
	"net/rpc"

	"github.com/korkmazkadir/rapidchain/consensus"
)

// Client implements P2P client
type P2PClient struct {
	IPAddress  string
	portNumber int

	rpcClient *rpc.Client

	blockChunks chan consensus.BlockChunk
	votes       chan consensus.Vote

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

	client.blockChunks = make(chan consensus.BlockChunk, 1024)
	client.votes = make(chan consensus.Vote, 1024)

	return client, nil
}

// Start starts the main loop of client. It blocks the calling goroutine
func (c *P2PClient) Start() {

	c.mainLoop()
}

// SendBlockChunk enques a chunk of a block to send
func (c *P2PClient) SendBlockChunk(chunk consensus.BlockChunk) {

	c.blockChunks <- chunk
}

// SendVote enques a vote to send
func (c *P2PClient) SendVote(vote consensus.Vote) {

	c.votes <- vote
}

func (c *P2PClient) mainLoop() {

	for {
		select {

		case blockChunk := <-c.blockChunks:
			c.rpcClient.Call("Server.HandleBlockChunk", blockChunk, nil)

		case vote := <-c.votes:
			c.rpcClient.Call("Server.HandleVote", vote, nil)

		}
	}
}
