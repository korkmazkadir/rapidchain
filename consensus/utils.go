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

func receiveMultipleBlocks(round int, demux *common.Demux, chunkCount int, peerSet *network.PeerSet, leaderCount int) ([]common.Block, [][]byte) {

	chunkChan, err := demux.GetVoteBlockChunkChan(round)
	if err != nil {
		panic(err)
	}

	receiver := newBlockReceiver(leaderCount, chunkCount)
	for !receiver.ReceivedAll() {
		c := <-chunkChan
		receiver.AddChunk(c)
		peerSet.ForwardChunk(c)
	}

	return receiver.GetBlocks()
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

func receiveMultipleProposeVotes(round int, demux *common.Demux, peerSet *network.PeerSet, leaderCount int) []common.Vote {

	proposeChannel, err := demux.GetVoteChan(round, common.ProposeTag)
	if err != nil {
		panic(err)
	}

	var proposeVotes []common.Vote
	for {

		vote := <-proposeChannel
		if !validateVote(vote, nil) {
			log.Printf("invalid propose vote recevied: %+v\n", vote)
			continue
		}

		peerSet.ForwardVote(vote)

		proposeVotes = append(proposeVotes, vote)
		if len(proposeVotes) == leaderCount {
			return sortProposeVotes(proposeVotes)
		}

	}

}

func sortProposeVotes(votes []common.Vote) []common.Vote {
	sort.Slice(votes, func(i, j int) bool {
		return bytes.Compare(votes[i].BlockHash[0], votes[j].BlockHash[0]) == -1
	})
	return votes
}

// I can count number of votes, and I can return error after receiving (f/2)+1 votes

func receiveEchoVotes(round int, demux *common.Demux, minVoteCount int, merkleRoots [][]byte, peerSet *network.PeerSet) []common.Vote {

	echoChannel, err := demux.GetVoteChan(round, common.EchoTag)
	if err != nil {
		panic(err)
	}

	var echoVotes []common.Vote

	for {

		ev := <-echoChannel

		if !AreTheyEqual(merkleRoots, ev.BlockHash) || !validateVote(ev, merkleRoots) {
			log.Printf("echo vore received for undefined merkleroot \n")
			continue
		}

		echoVotes = append(echoVotes, ev)
		peerSet.ForwardVote(ev)

		if len(echoVotes) == minVoteCount {
			return echoVotes
		}

	}

}

func AreTheyEqual(merkleRoots1 [][]byte, merkleRoots2 [][]byte) bool {
	if len(merkleRoots1) != len(merkleRoots2) {
		return false
	}

	for i := range merkleRoots1 {
		if !bytes.Equal(merkleRoots1[i], merkleRoots2[i]) {
			return false
		}
	}

	return true
}

func receiveAcceptVotes(round int, demux *common.Demux, minVoteCount int, merkleRoots [][]byte, peerSet *network.PeerSet) []common.Vote {

	acceptChannel, err := demux.GetVoteChan(round, common.AcceptTag)
	if err != nil {
		panic(err)
	}

	var acceptVotes []common.Vote

	for {

		av := <-acceptChannel

		if !AreTheyEqual(merkleRoots, av.BlockHash) || len(av.Proof.EchoVotes) < minVoteCount {
			continue
		}

		isEchoVotesValid := validateVote(av, merkleRoots)
		for i := range av.Proof.EchoVotes {
			isEchoVotesValid = isEchoVotesValid && validateVote(av.Proof.EchoVotes[i], merkleRoots)
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

func validateVote(vote common.Vote, merkleRoots [][]byte) bool {

	return true
}

func signHash(hash []byte, keyPrive ed25519.PrivateKey) []byte {

	return ed25519.Sign(keyPrive, hash)
}

func encodeBase64(hex []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(hex))
}
