package consensus

import (
	"fmt"
	"sort"

	"github.com/korkmazkadir/rapidchain/common"
)

type blockReceiver struct {
	blockCount int
	chunkCount int
	blockMap   map[string][]common.BlockChunk
}

func newBlockReceiver(leaderCount int, chunkCount int) *blockReceiver {

	r := &blockReceiver{
		blockCount: leaderCount,
		chunkCount: chunkCount,
		blockMap:   make(map[string][]common.BlockChunk),
	}

	return r
}

// AddChunk stores a chunk of a block to reconstruct the whole block later
func (r *blockReceiver) AddChunk(chunk common.BlockChunk) {
	key := string(chunk.Authenticator.MerkleRoot)
	chunkSlice := r.blockMap[key]
	r.blockMap[key] = append(chunkSlice, chunk)

	if len(r.blockMap) > r.blockCount {
		panic(fmt.Errorf("there are more blocks than expected, the number of blocks is %d", len(r.blockMap)))
	}
}

// ReceivedAll checks whether all chunks are recived or not to reconstruct the blocks of a round
func (r *blockReceiver) ReceivedAll() bool {

	if len(r.blockMap) != r.blockCount {
		return false
	}

	for _, chunkSlice := range r.blockMap {
		if len(chunkSlice) != r.chunkCount {
			return false
		}
	}

	return true
}

// GetBlocks recunstruct blocks using chunks, and returns the blocks by sorting the resulting block slice according to block hashes
func (r *blockReceiver) GetBlocks() ([]common.Block, [][]byte) {

	if r.ReceivedAll() == false {
		panic(fmt.Errorf("not received all block chunks to reconstruct block/s"))
	}

	keys := make([]string, 0, len(r.blockMap))
	for k := range r.blockMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var blocks []common.Block
	var merkleRoots [][]byte
	for _, key := range keys {

		merkleRoots = append(merkleRoots, []byte(key))

		receivedChunks := r.blockMap[key]
		sort.Slice(receivedChunks, func(i, j int) bool {
			return receivedChunks[i].ChunkIndex < receivedChunks[j].ChunkIndex
		})

		block := common.MergeChunks(receivedChunks)
		blocks = append(blocks, block)
	}

	return blocks, merkleRoots
}
