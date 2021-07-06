package consensus

import (
	"bytes"
	"errors"

	"github.com/korkmazkadir/rapidchain/common"
	"github.com/korkmazkadir/rapidchain/network"
	"github.com/korkmazkadir/rapidchain/registry"
)

// BlockNotValid is returned if the block can not pass vaslidity test
var ErrBlockNotValid = errors.New("received block is not valid")

// BlockNotValid is returned if the block can not pass vaslidity test
var ErrDecidedOnDifferentBlock = errors.New("decided on a different block, possibly the leader equivocate")

type RapidchainConsensus struct {
	demultiplexer *common.Demux
	nodeConfig    registry.NodeConfig
	peerSet       network.PeerSet

	keyPublick []byte
	keySecret  []byte
}

func (c *RapidchainConsensus) Propose(round int, block common.Block, previousBlockHash []byte) (common.Block, error) {

	// sets the round for demultiplexer
	c.demultiplexer.UpdateRound(round)

	// chunks the block
	chunks, merkleRoot := common.ChunkBlock(block, c.nodeConfig.BlockChunkCount)

	// disseminate chunks over different nodes
	c.peerSet.DissaminateChunks(chunks)

	// vote propose
	c.vote(common.ProposeTag, round, merkleRoot, nil)

	// I am returning accept votes but I do not know how to use them!!!!
	c.commonPath(round, merkleRoot)

	return block, nil
}

func (c *RapidchainConsensus) Decide(round int, previousBlockHash []byte) (common.Block, error) {
	// sets the round for demultiplexer
	c.demultiplexer.UpdateRound(round)

	proposeVote := receiveProposeVote(round, c.demultiplexer, &c.peerSet)

	// receives the block
	block, merkleRoot, err := receiveBlock(round, c.demultiplexer, c.nodeConfig.BlockChunkCount, &c.peerSet)
	if err != nil {
		return common.Block{}, err
	}

	valid := validateBlock(block, previousBlockHash)
	if !valid {
		return common.Block{}, ErrBlockNotValid
	}

	if !bytes.Equal(proposeVote.BlockHash, merkleRoot) {
		return common.Block{}, ErrDecidedOnDifferentBlock
	}

	// I am returning accept votes but I do not know how to use them!!!!
	c.commonPath(round, merkleRoot)

	return block, nil
}

func (c *RapidchainConsensus) commonPath(round int, merkleRoot []byte) []common.Vote {

	// vote echo
	c.vote(common.EchoTag, round, merkleRoot, nil)

	minVoteCount := (c.nodeConfig.NodeCount / 2) + 1
	echoVotes := receiveEchoVotes(round, c.demultiplexer, minVoteCount, merkleRoot, &c.peerSet)

	acceptProof := common.AcceptProof{EchoVotes: echoVotes}
	c.vote(common.AcceptTag, round, merkleRoot, &acceptProof)

	acceptVotes := receiveAcceptVotes(round, c.demultiplexer, minVoteCount, merkleRoot, &c.peerSet)

	// I am returning accept votes but I do not know how to use them!!!!
	return acceptVotes
}

func (c *RapidchainConsensus) vote(tag byte, round int, merkleRoot []byte, proof *common.AcceptProof) {
	vote := common.Vote{
		Issuer:    c.keyPublick,
		Tag:       tag,
		Round:     round,
		BlockHash: merkleRoot,
	}

	if proof != nil {
		vote.Proof = *proof
	}

	vote.Signature = signHash(vote.Hash(), c.keySecret)

	c.peerSet.ForwardVote(vote)
}
