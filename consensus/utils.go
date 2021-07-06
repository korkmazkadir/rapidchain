package consensus

import (
	"bytes"
	"log"

	"github.com/korkmazkadir/rapidchain/common"
	"github.com/korkmazkadir/rapidchain/network"
)

func receiveBlock(round int, demux *common.Demux, chunkCount int, peerSet *network.PeerSet) (common.Block, []byte, error) {

	return common.Block{}, []byte{}, nil
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
			continue
		}

		echoVotes = append(echoVotes, ev)
		peerSet.ForwardVote(ev)

		if len(echoVotes) >= minVoteCount {
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

func signHash(hash []byte, keySecret []byte) []byte {
	//panic("signVote not implemented")
	return []byte{}
}
