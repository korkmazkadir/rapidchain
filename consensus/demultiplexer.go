package consensus

import (
	"sync"
)

type Demux struct {
	mutex sync.Mutex

	currentRound int

	voteChanMap map[int]chan Vote

	blockChanMap map[int]chan BlockChunk
}

func (d *Demux) EnqueBlockChunk(chunk BlockChunk) {

	d.mutex.Lock()
	defer d.mutex.Unlock()
}

func (d *Demux) EnqueVote(vote Vote) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

}

func (d *Demux) UpdateRound(round int) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

}
