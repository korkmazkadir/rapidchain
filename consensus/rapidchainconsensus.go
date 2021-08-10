package consensus

import (
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/korkmazkadir/rapidchain/common"
	"github.com/korkmazkadir/rapidchain/network"
	"github.com/korkmazkadir/rapidchain/registery"
)

// BlockNotValid is returned if the block can not pass vaslidity test
var ErrBlockNotValid = errors.New("received block is not valid")

// BlockNotValid is returned if the block can not pass vaslidity test
var ErrDecidedOnDifferentBlock = errors.New("decided on a different block, possibly the leader equivocate")

type RapidchainConsensus struct {
	demultiplexer *common.Demux
	nodeConfig    registery.NodeConfig
	peerSet       *network.PeerSet

	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey

	statLogger *common.StatLogger
}

func NewRapidchain(demux *common.Demux, config registery.NodeConfig, peerSet *network.PeerSet, statLogger *common.StatLogger) *RapidchainConsensus {

	keyPub, keyPrive, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	rapidchain := &RapidchainConsensus{
		demultiplexer: demux,
		nodeConfig:    config,
		peerSet:       peerSet,
		publicKey:     keyPub,
		privateKey:    keyPrive,
		statLogger:    statLogger,
	}

	return rapidchain
}

func (c *RapidchainConsensus) Propose(round int, block common.Block, previousBlockHash []byte) []common.Block {

	// starts a new epoch
	c.statLogger.NewRound(round)

	// sets the round for demultiplexer
	c.demultiplexer.UpdateRound(round)

	// chunks the block
	chunks := common.ChunkBlock(block, c.nodeConfig.BlockChunkCount)
	//log.Printf("proposing block %x\n", encodeBase64(merkleRoot[:15]))

	// signs chunks
	for i := range chunks {
		chunks[i].Issuer = c.publicKey
	}

	// disseminate chunks over different nodes
	c.peerSet.DissaminateChunks(chunks)

	return c.commonPath(round, previousBlockHash)
}

func (c *RapidchainConsensus) Decide(round int, previousBlockHash []byte) []common.Block {

	// starts a new epoch
	c.statLogger.NewRound(round)

	// sets the round for demultiplexer
	c.demultiplexer.UpdateRound(round)

	return c.commonPath(round, previousBlockHash)
}

func (c *RapidchainConsensus) commonPath(round int, previousBlockHash []byte) []common.Block {

	// BLOCK RECEIVE EVENT
	//log.Printf("waiting for block...\n")
	startTime := time.Now()
	blocks := receiveMultipleBlocks(round, c.demultiplexer, c.nodeConfig.BlockChunkCount, c.peerSet, c.nodeConfig.LeaderCount)
	c.statLogger.LogBlockReceive(time.Since(startTime).Milliseconds())

	c.statLogger.LogEndOfRound()

	// I am returning accept votes but I do not know how to use them!!!!
	return blocks
}
