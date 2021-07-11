package common

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"
)

func TestErasurecode(t *testing.T) {

	currentRound := 1

	block := Block{
		Round:         currentRound,
		Issuer:        getRandomByteSlice(32),
		Payload:       getRandomByteSlice(2097152),
		PrevBlockHash: getRandomByteSlice(32),
	}

	chunkCount := 128
	chunks, _ := ChunkBlock(block, chunkCount)

	dataShardCount := 16
	parityShardCount := 3

	startTime := time.Now()
	chunksCoded := ErasureCode(currentRound, block.Issuer, chunks, dataShardCount, parityShardCount)
	elapsedTime := time.Since(startTime).Milliseconds()

	t.Logf("Elapsed time to erasurecode %d", elapsedTime)
	t.Logf("Len(chunks) %d", len(chunks))
	t.Logf("Len(chunksCoded) %d", len(chunksCoded))

	result := AreThereEnoughChunksToReconstructBlock(chunksCoded, dataShardCount, parityShardCount)

	t.Logf("Are there enough chunks %t", result)

	// removes a chunk
	chunksCoded[0].Payload = nil

	startTime = time.Now()
	constructedChunks := ReconstructMissingChunks(chunksCoded, dataShardCount, parityShardCount)
	elapsedTime = time.Since(startTime).Milliseconds()

	t.Logf("Elapsed time to reconstruct missing chunks %d", elapsedTime)

	h := sha256.New()
	h.Write(chunks[0].Payload)
	hash1 := h.Sum(nil)
	t.Logf("Chunks[0] payload %x", hash1)

	h.Reset()
	h.Write(constructedChunks[0].Payload)
	hash2 := h.Sum(nil)
	t.Logf("constructedChunks[0] payload %x", hash2)

	if !bytes.Equal(hash1, hash2) {
		t.Errorf("cound not reconbstruct the original chunk expected [%x] reconstrcuted [%x] ", hash1, hash2)
	}

}
