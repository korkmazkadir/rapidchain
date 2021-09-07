package consensus

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/korkmazkadir/rapidchain/common"
)

type blockReceiver struct {
	blockCount     int
	chunkCount     int
	blockMap       map[string][]common.BlockChunk
	wg             sync.WaitGroup
	receivedBlocks map[string]common.Block
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

	if len(r.blockMap[key]) == r.chunkCount {
		// it means that we have the all chunks of the microblock
		// we can walidate it here
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()

			receivedChunks := r.blockMap[key]
			sort.Slice(receivedChunks, func(i, j int) bool {
				return receivedChunks[i].ChunkIndex < receivedChunks[j].ChunkIndex
			})

			block := common.MergeChunks(receivedChunks)
			log.Printf("[%s] chunked count of the recived block is %d payload is %d bytes\n", encodeBase64([]byte(key[:15])), len(receivedChunks), len(block.Payload))
			r.receivedBlocks[key] = block
			payloadSize := len(block.Payload)

			// emulating cost of validating a micro block
			sleepTime := (float64(0.133) * float64(payloadSize/512))
			sleepDuration := time.Duration(sleepTime) * time.Millisecond
			log.Printf("the node will sleep to emulate tx validation, and merkle tree construction %s \n", sleepDuration)
			time.Sleep(sleepDuration)

		}()

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

	// waiting for the block validation before returning blocks
	r.wg.Wait()

	keys := make([]string, 0, len(r.receivedBlocks))
	for k := range r.receivedBlocks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var blocks []common.Block
	var merkleRoots [][]byte
	for _, key := range keys {

		merkleRoots = append(merkleRoots, []byte(key))
		blocks = append(blocks, r.receivedBlocks[key])
	}

	return blocks, merkleRoots
}
