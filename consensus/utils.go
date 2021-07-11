package consensus

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"log"
	"sort"

	"github.com/korkmazkadir/rapidchain/common"
	"github.com/korkmazkadir/rapidchain/network"
)

// receiveBlock returns block, merkle root, error
func receiveBlock(round int, demux *common.Demux, chunkCount int, peerSet *network.PeerSet) (common.Block, []byte, error) {

	chunkChan, err := demux.GetVoteBlockChunkChan(round)
	if err != nil {
		panic(err)
	}

	// check for differet merkle roots and return error
	var receivedChunks []common.BlockChunk
	for len(receivedChunks) < chunkCount {
		c := <-chunkChan
		receivedChunks = append(receivedChunks, c)
		peerSet.ForwardChunk(c)
	}

	sort.Slice(receivedChunks, func(i, j int) bool {
		return receivedChunks[i].ChunkIndex < receivedChunks[j].ChunkIndex
	})

	block := common.MergeChunks(receivedChunks)

	// this way of returnin merkleroot is wrong
	return block, receivedChunks[0].Authenticator.MerkleRoot, nil
}

func receiveErasureCodedBlock(round int, demux *common.Demux, chunkCount int, peerSet *network.PeerSet, dataChunkCount int, parityChunkCount int) (common.Block, []byte, error) {

	chunkChan, err := demux.GetVoteBlockChunkChan(round)
	if err != nil {
		panic(err)
	}

	// check for differet merkle roots and return error
	ck := chunkCount + ((chunkCount / dataChunkCount) * parityChunkCount)
	receivedChunks := make([]common.BlockChunk, ck)
	var merkleRoot []byte
	for common.AreThereEnoughChunksToReconstructBlock(receivedChunks, dataChunkCount, parityChunkCount) == false {
		c := <-chunkChan
		receivedChunks[c.ChunkIndex] = c
		peerSet.ForwardChunk(c)

		//
		// FIX this!!!
		// This is not good way to get merkle root
		// Consider waiting for a specific block by getting merkle root!!!
		if len(merkleRoot) == 0 {
			merkleRoot = append(merkleRoot, c.Authenticator.MerkleRoot...)
		}

	}

	block := common.MergeErasureCodedChunks(receivedChunks, dataChunkCount, parityChunkCount)

	log.Printf("Received block hash is %x\n", block.Hash()[:16])

	// this way of returnin merkleroot is wrong
	return block, merkleRoot, nil
}

func receiveProposeVote(round int, demux *common.Demux, peerSet *network.PeerSet) common.Vote {

	proposeChannel, err := demux.GetVoteChan(round, common.ProposeTag)
	if err != nil {
		panic(err)
	}

	var proposeVote common.Vote
	for {

		proposeVote = <-proposeChannel
		if !validateVote(proposeVote, nil) {
			log.Printf("invalid propose vote recevied: %+v\n", proposeVote)
			continue
		}

		peerSet.ForwardVote(proposeVote)
		return proposeVote
	}

}

// I can count number of votes, and I can return error after receiving (f/2)+1 votes

func receiveEchoVotes(round int, demux *common.Demux, minVoteCount int, merkleRoot []byte, peerSet *network.PeerSet) []common.Vote {

	echoChannel, err := demux.GetVoteChan(round, common.EchoTag)
	if err != nil {
		panic(err)
	}

	var echoVotes []common.Vote

	for {

		ev := <-echoChannel

		if !bytes.Equal(ev.BlockHash, merkleRoot) || !validateVote(ev, merkleRoot) {
			log.Printf("echo vore received for undefined merkleroot %x\n", encodeBase64(ev.BlockHash))
			continue
		}

		echoVotes = append(echoVotes, ev)
		peerSet.ForwardVote(ev)

		if len(echoVotes) == minVoteCount {
			return echoVotes
		}

	}

}

func receiveAcceptVotes(round int, demux *common.Demux, minVoteCount int, merkleRoot []byte, peerSet *network.PeerSet) []common.Vote {

	acceptChannel, err := demux.GetVoteChan(round, common.AcceptTag)
	if err != nil {
		panic(err)
	}

	var acceptVotes []common.Vote

	for {

		av := <-acceptChannel

		if !bytes.Equal(av.BlockHash, merkleRoot) || len(av.Proof.EchoVotes) < minVoteCount {
			continue
		}

		isEchoVotesValid := validateVote(av, merkleRoot)
		for i := range av.Proof.EchoVotes {
			isEchoVotesValid = isEchoVotesValid && validateVote(av.Proof.EchoVotes[i], merkleRoot)
			if !isEchoVotesValid {
				break
			}
		}

		if !isEchoVotesValid {
			continue
		}

		acceptVotes = append(acceptVotes, av)
		peerSet.ForwardVote(av)

		if len(acceptVotes) >= minVoteCount {
			return acceptVotes
		}

	}
}

func validateBlock(block common.Block, previousBlockHash []byte) bool {

	return true
}

func validateVote(vote common.Vote, merkleRoot []byte) bool {

	return true
}

func signHash(hash []byte, keyPrive ed25519.PrivateKey) []byte {

	return ed25519.Sign(keyPrive, hash)
}

func encodeBase64(hex []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(hex))
}
