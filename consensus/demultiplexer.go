package consensus

import (
	"fmt"
	"sync"
)

const (
	channelCapacity = 1024
)

// Demux provides message multiplexing service
// Network and consensus layer communicate using demux
type Demux struct {
	mutex sync.Mutex

	currentRound int

	// it is used to filter already processed messages
	processedMessageMap map[int]map[string]struct{}

	proposeVoteChanMap map[int]chan Vote

	echoVoteChanMap map[int]chan Vote

	acceptVoteChanMap map[int]chan Vote

	blockChunkChanMap map[int]chan BlockChunk
}

// NewDemultiplexer creates a new demultiplexer with initial round value
func NewDemultiplexer(initialRound int) *Demux {

	demux := &Demux{currentRound: initialRound}

	demux.processedMessageMap = make(map[int]map[string]struct{})
	demux.proposeVoteChanMap = make(map[int]chan Vote)
	demux.echoVoteChanMap = make(map[int]chan Vote)
	demux.acceptVoteChanMap = make(map[int]chan Vote)
	demux.blockChunkChanMap = make(map[int]chan BlockChunk)

	return demux
}

// EnqueBlockChunk enques a block chunk to be the consumed by consensus layer
func (d *Demux) EnqueBlockChunk(chunk BlockChunk) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if chunk.Round < d.currentRound {
		// discarts a chunks because it belongs to a previous round
		return
	}

	chunkRound := chunk.Round
	chunkHash := string(chunk.Hash())
	if d.isProcessed(chunkRound, chunkHash) {
		// chunk is already processed
		return
	}

	chunkChan := d.getCorrespondingBlockChunkChan(chunkRound)
	chunkChan <- chunk

	d.markAsProcessed(chunkRound, chunkHash)
}

// EnqueVote enques a vote to be consumed by the consensus layer
func (d *Demux) EnqueVote(vote Vote) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if vote.Round < d.currentRound {
		// discarts a round because it belongs to a previous round
		return
	}

	voteRound := vote.Round
	voteHash := string(vote.Hash())
	if d.isProcessed(voteRound, voteHash) {
		// vote is already processed
		return
	}

	voteChan := d.getCorrespondingVoteChan(vote.Round, vote.Tag)
	voteChan <- vote

	d.markAsProcessed(voteRound, voteHash)
}

// GetVoteChan returns vote channel
func (d *Demux) GetVoteChan(round int, tag byte) (chan Vote, error) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if round < d.currentRound {
		return nil, fmt.Errorf("the current round value is bigger than the provided round value")
	}

	return d.getCorrespondingVoteChan(round, tag), nil
}

// GetVoteBlockChunkChan returns Blockchunk channel
func (d *Demux) GetVoteBlockChunkChan(round int) (chan BlockChunk, error) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if round < d.currentRound {
		return nil, fmt.Errorf("the current round value is bigger than the provided round value")
	}

	return d.getCorrespondingBlockChunkChan(round), nil
}

// UpdateRound updates the round.
// All messages blongs to the previous rounds discarted
// Update round mustbe called by an increased round number otherwise this function panics
func (d *Demux) UpdateRound(round int) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Round value should increase one by one
	if round < d.currentRound || round != (d.currentRound+1) {
		panic(fmt.Errorf("illegal round value, current round value %d, provided round value %d", d.currentRound, round))
	}

	d.currentRound = round
	d.deletePreviousRoundMessages()
}

// All the following functions are helper functions.
// They must be called from previous functions because
// they are not thread safe!

func (d *Demux) deletePreviousRoundMessages() {

	previousRound := d.currentRound - 1

	delete(d.processedMessageMap, previousRound)
	delete(d.proposeVoteChanMap, previousRound)
	delete(d.echoVoteChanMap, previousRound)
	delete(d.acceptVoteChanMap, previousRound)
	delete(d.blockChunkChanMap, previousRound)

}

func (d *Demux) getProcessedMessageMap(round int) map[string]struct{} {

	if val, ok := d.processedMessageMap[round]; ok {
		return val
	}

	val := make(map[string]struct{})
	d.processedMessageMap[round] = val

	return val
}

func (d *Demux) isProcessed(round int, hash string) bool {

	processedMessageMap := d.getProcessedMessageMap(round)
	chunkHashString := string(hash)
	_, ok := processedMessageMap[chunkHashString]
	return ok
}

func (d *Demux) markAsProcessed(round int, hash string) {

	processedMessageMap := d.getProcessedMessageMap(round)
	processedMessageMap[hash] = struct{}{}
}

func (d *Demux) getCorrespondingVoteChan(round int, tag byte) chan Vote {

	var correspondingVoteMap map[int]chan Vote

	switch tag {
	case ProposeTag:
		correspondingVoteMap = d.proposeVoteChanMap
		break
	case EchoTag:
		correspondingVoteMap = d.echoVoteChanMap
		break

	case AcceptTag:
		correspondingVoteMap = d.acceptVoteChanMap
		break
	default:
		panic(fmt.Errorf("unknown tag received %b", tag))
	}

	if val, ok := correspondingVoteMap[round]; ok {
		return val
	}

	voteChan := make(chan Vote, channelCapacity)
	correspondingVoteMap[round] = voteChan

	return voteChan
}

func (d *Demux) getCorrespondingBlockChunkChan(round int) chan BlockChunk {

	if val, ok := d.blockChunkChanMap[round]; ok {
		return val
	}

	chunkChan := make(chan BlockChunk, channelCapacity)
	d.blockChunkChanMap[round] = chunkChan

	return chunkChan
}
