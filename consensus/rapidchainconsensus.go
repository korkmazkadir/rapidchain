package consensus

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"log"
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
	peerSet       network.PeerSet

	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey

	statLogger *common.StatLogger
}

func NewRapidchain(demux *common.Demux, config registery.NodeConfig, peerSet network.PeerSet, statLogger *common.StatLogger) *RapidchainConsensus {

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
	chunks, merkleRoot := common.ChunkBlock(block, c.nodeConfig.BlockChunkCount)
	//log.Printf("proposing block %x\n", encodeBase64(merkleRoot[:15]))
	log.Printf("the block chunked into %d chunks \n", len(chunks))

	// signs chunks
	for i := range chunks {
		chunks[i].Issuer = c.publicKey
		chunks[i].Signature = signHash(chunks[i].Hash(), c.privateKey)
	}

	// disseminate chunks over different nodes
	c.peerSet.DissaminateChunks(chunks)

	// vote propose
	c.vote(common.ProposeTag, round, [][]byte{merkleRoot}, nil)

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

	// PROPOSE EVENT
	startTime := time.Now()
	proposeVotes := receiveMultipleProposeVotes(round, c.demultiplexer, &c.peerSet, c.nodeConfig.LeaderCount)
	c.statLogger.LogPropose(time.Since(startTime).Milliseconds())

	// BLOCK RECEIVE EVENT
	//log.Printf("waiting for block...\n")
	startTime = time.Now()
	blocks, merkleRoots := receiveMultipleBlocks(round, c.demultiplexer, c.nodeConfig.BlockChunkCount, &c.peerSet, c.nodeConfig.LeaderCount)
	c.statLogger.LogBlockReceive(time.Since(startTime).Milliseconds())

	for _, block := range blocks {
		valid := validateBlock(block, previousBlockHash)
		if !valid {
			panic(ErrBlockNotValid)
		}
	}

	for i := range proposeVotes {
		if !bytes.Equal(proposeVotes[i].BlockHash[0], merkleRoots[i]) {
			panic(ErrDecidedOnDifferentBlock)
		}
	}

	// vote echo
	c.vote(common.EchoTag, round, merkleRoots, nil)

	// ECHO EVENT
	minVoteCount := (c.nodeConfig.NodeCount / 2) + 1
	//log.Printf("waiting for %d echoes \n", minVoteCount)
	startTime = time.Now()
	echoVotes := receiveEchoVotes(round, c.demultiplexer, minVoteCount, merkleRoots, &c.peerSet)
	c.statLogger.LogEcho(time.Since(startTime).Milliseconds())

	acceptProof := common.AcceptProof{EchoVotes: echoVotes}
	c.vote(common.AcceptTag, round, merkleRoots, &acceptProof)

	// ACCEPT EVENT
	//log.Printf("waiting for %d accept \n", minVoteCount)
	//startTime = time.Now()
	//receiveAcceptVotes(round, c.demultiplexer, minVoteCount, merkleRoots, &c.peerSet)
	//c.statLogger.LogAccept(time.Since(startTime).Milliseconds())

	c.statLogger.LogEndOfRound()

	// I am returning accept votes but I do not know how to use them!!!!
	return blocks
}

func (c *RapidchainConsensus) vote(tag byte, round int, merkleRoots [][]byte, proof *common.AcceptProof) {
	vote := common.Vote{
		Issuer:    c.publicKey,
		Tag:       tag,
		Round:     round,
		BlockHash: merkleRoots,
	}

	if proof != nil {
		vote.Proof = *proof
	}

	vote.Signature = signHash(vote.Hash(), c.privateKey)

	c.peerSet.ForwardVote(vote)
}
