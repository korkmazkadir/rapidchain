package common

import (
	"testing"
)

func TestDemultiplexer(t *testing.T) {

	currentRound := 1
	demux := NewDemultiplexer(currentRound)

	block := Block{
		Round:         currentRound,
		Issuer:        getRandomByteSlice(32),
		Payload:       getRandomByteSlice(2097152),
		PrevBlockHash: getRandomByteSlice(32),
	}

	chunkCount := 128
	chunks, _ := ChunkBlock(block, chunkCount)

	for i := range chunks {
		demux.EnqueBlockChunk(chunks[i])
	}

	// try to reenque, it is not allowed
	for i := range chunks {
		demux.EnqueBlockChunk(chunks[i])
	}

	chunkChan, err := demux.GetVoteBlockChunkChan(currentRound)

	if err != nil {
		t.Error(err)
	}

	if chunkChan == nil {
		t.Errorf("returned channel is nil")
	}

	if len(chunkChan) != chunkCount {
		t.Errorf("expected chunk count is %d received %d chunk", chunkCount, len(chunkChan))
	}

	demux.UpdateRound(2)
	chunkChan, err = demux.GetVoteBlockChunkChan(currentRound)

	if err == nil {
		t.Errorf("expecting non nil error because try to access the previous round value")
	}

	currentRound = 2
	chunkChan, err = demux.GetVoteBlockChunkChan(currentRound)

	if err != nil {
		t.Error(err)
	}

	// Propose vote

	voteCount := 32
	for i := 0; i < 32; i++ {
		v := Vote{
			Issuer:    getRandomByteSlice(32),
			Tag:       ProposeTag,
			Round:     currentRound,
			BlockHash: getRandomByteSlice(32),
		}

		demux.EnqueVote(v)
		demux.EnqueVote(v)
		demux.EnqueVote(v)
	}

	proposeVoteChan, err := demux.GetVoteChan(currentRound, ProposeTag)
	if err != nil {
		t.Error(err)
	}

	if len(proposeVoteChan) != voteCount {
		t.Errorf("expected propose vote count is %d received %d propose vote", voteCount, len(proposeVoteChan))
	}

	for i := 0; i < 32; i++ {
		v := Vote{
			Issuer:    getRandomByteSlice(32),
			Tag:       EchoTag,
			Round:     currentRound,
			BlockHash: getRandomByteSlice(32),
		}

		demux.EnqueVote(v)
		demux.EnqueVote(v)
		demux.EnqueVote(v)
	}

	echoVoteChan, err := demux.GetVoteChan(currentRound, EchoTag)
	if err != nil {
		t.Error(err)
	}

	if len(echoVoteChan) != voteCount {
		t.Errorf("expected propose vote count is %d received %d propose vote", voteCount, len(echoVoteChan))
	}

	acceptVoteChan, err := demux.GetVoteChan(currentRound, AcceptTag)
	if err != nil {
		t.Error(err)
	}

	if len(acceptVoteChan) != 0 {
		t.Errorf("expected vote count is 0, received ,%d", len(acceptVoteChan))
	}

	for i := 0; i < 32; i++ {
		v := Vote{
			Issuer:    getRandomByteSlice(32),
			Tag:       AcceptTag,
			Round:     currentRound,
			BlockHash: getRandomByteSlice(32),
		}

		// only sigle value will be enqueued
		demux.EnqueVote(v)
		demux.EnqueVote(v)
		demux.EnqueVote(v)
	}

	if len(acceptVoteChan) != voteCount {
		t.Errorf("eexpected vote count is %d, received ,%d", voteCount, len(acceptVoteChan))
	}

}
