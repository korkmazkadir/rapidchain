package registery

import (
	"crypto/sha256"
	"fmt"
)

type NodeConfig struct {
	NodeCount int

	EpochSeed []byte

	EndRound int

	GossipFanout int

	BlockSize int

	BlockChunkCount int

	DataChunkCount int

	PairtyChunkCount int
}

func (nc NodeConfig) Hash() []byte {

	str := fmt.Sprintf("%d,%x,%d,%d,%d,%d,%d,%d", nc.NodeCount, nc.EpochSeed, nc.EndRound, nc.GossipFanout, nc.BlockSize, nc.BlockChunkCount,
		nc.DataChunkCount, nc.PairtyChunkCount)

	h := sha256.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

func (nc *NodeConfig) CopyFields(cp NodeConfig) {
	nc.NodeCount = cp.NodeCount
	nc.EpochSeed = nc.EpochSeed[:0]
	nc.EpochSeed = append(nc.EpochSeed, cp.EpochSeed...)
	nc.EndRound = cp.EndRound
	nc.GossipFanout = cp.GossipFanout
	nc.BlockSize = cp.BlockSize
	nc.BlockChunkCount = cp.BlockChunkCount
	nc.DataChunkCount = cp.DataChunkCount
	nc.PairtyChunkCount = cp.PairtyChunkCount
}
