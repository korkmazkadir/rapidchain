package consensus

import (
	"bytes"
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

func (c *RapidchainConsensus) Propose(round int, block common.Block, previousBlockHash []byte) (common.Block, error) {

	// starts a new epoch
	c.statLogger.NewRound(round)

	// sets the round for demultiplexer
	c.demultiplexer.UpdateRound(round)

	// chunks the block
	chunks, merkleRoot := common.ChunkBlock(block, c.nodeConfig.BlockChunkCount)
	//log.Printf("proposing block %x\n", encodeBase64(merkleRoot[:15]))

	// disseminate chunks over different nodes
	c.peerSet.DissaminateChunks(chunks)

	// vote propose
	c.vote(common.ProposeTag, round, merkleRoot, nil)

	// I am returning accept votes but I do not know how to use them!!!!
	c.commonPath(round, merkleRoot)

	return block, nil
}

func (c *RapidchainConsensus) Decide(round int, previousBlockHash []byte) (common.Block, error) {

	// starts a new epoch
	c.statLogger.NewRound(round)

	// sets the round for demultiplexer
	c.demultiplexer.UpdateRound(round)

	//log.Printf("waiting for propose...\n")

	// PROPOSE EVENT
	startTime := time.Now()
	proposeVote := receiveProposeVote(round, c.demultiplexer, &c.peerSet)
	c.statLogger.LogPropose(time.Since(startTime).Milliseconds())

	// BLOCK RECEIVE EVENT
	//log.Printf("waiting for block...\n")
	startTime = time.Now()
	block, merkleRoot, err := receiveBlock(round, c.demultiplexer, c.nodeConfig.BlockChunkCount, &c.peerSet)
	if err != nil {
		return common.Block{}, err
	}
	c.statLogger.LogBlockReceive(time.Since(startTime).Milliseconds())

	valid := validateBlock(block, previousBlockHash)
	if !valid {
		return common.Block{}, ErrBlockNotValid
	}

	if !bytes.Equal(proposeVote.BlockHash, merkleRoot) {
		return common.Block{}, ErrDecidedOnDifferentBlock
	}

	// I am returning accept votes but I do not know how to use them!!!!
	//log.Printf("---Common Path---\n")
	c.commonPath(round, merkleRoot)

	return block, nil
}

func (c *RapidchainConsensus) commonPath(round int, merkleRoot []byte) []common.Vote {

	// vote echo
	c.vote(common.EchoTag, round, merkleRoot, nil)

	// ECHO EVENT
	minVoteCount := (c.nodeConfig.NodeCount / 2) + 1
	//log.Printf("waiting for %d echoes \n", minVoteCount)
	startTime := time.Now()
	echoVotes := receiveEchoVotes(round, c.demultiplexer, minVoteCount, merkleRoot, &c.peerSet)
	c.statLogger.LogEcho(time.Since(startTime).Milliseconds())

	acceptProof := common.AcceptProof{EchoVotes: echoVotes}
	c.vote(common.AcceptTag, round, merkleRoot, &acceptProof)

	// ACCEPT EVENT
	//log.Printf("waiting for %d accept \n", minVoteCount)
	startTime = time.Now()
	acceptVotes := receiveAcceptVotes(round, c.demultiplexer, minVoteCount, merkleRoot, &c.peerSet)
	c.statLogger.LogAccept(time.Since(startTime).Milliseconds())

	c.statLogger.LogEndOfRound()

	// I am returning accept votes but I do not know how to use them!!!!
	return acceptVotes
}

func (c *RapidchainConsensus) vote(tag byte, round int, merkleRoot []byte, proof *common.AcceptProof) {
	vote := common.Vote{
		Issuer:    c.publicKey,
		Tag:       tag,
		Round:     round,
		BlockHash: merkleRoot,
	}

	if proof != nil {
		vote.Proof = *proof
	}

	vote.Signature = signHash(vote.Hash(), c.privateKey)

	c.peerSet.ForwardVote(vote)
}
