package consensus

import (
	"encoding/base64"

	"github.com/korkmazkadir/rapidchain/common"
	"github.com/korkmazkadir/rapidchain/network"
)

func receiveMultipleBlocks(round int, demux *common.Demux, chunkCount int, peerSet *network.PeerSet, leaderCount int) []common.Block {

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

func encodeBase64(hex []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(hex))
}
