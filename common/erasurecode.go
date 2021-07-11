package common

import (
	"github.com/klauspost/reedsolomon"
)

func ErasureCode(round int, issuer []byte, chunks []BlockChunk, dataShardCount int, parityShardCount int) []BlockChunk {

	var erasureCodedChunks []BlockChunk
	for i := 0; i < len(chunks); i += dataShardCount {
		c := chunks[i : i+dataShardCount]
		sharded := addParityShards(c, dataShardCount, parityShardCount)
		erasureCodedChunks = append(erasureCodedChunks, sharded...)
	}

	finalChunkCount := len(erasureCodedChunks)
	for i := range erasureCodedChunks {
		erasureCodedChunks[i].Issuer = issuer
		erasureCodedChunks[i].Round = round
		erasureCodedChunks[i].ChunkIndex = i
		erasureCodedChunks[i].ChunkCount = finalChunkCount
		erasureCodedChunks[i].ErorrCorrectionScheme[0] = dataShardCount
		erasureCodedChunks[i].ErorrCorrectionScheme[1] = parityShardCount
	}

	return erasureCodedChunks
}

// AreThereEnoughChunksToReconstructBlock assumes that all chunks belongs to same block
func AreThereEnoughChunksToReconstructBlock(chunks []BlockChunk, dataShardCount int, parityShardCount int) bool {

	shardSize := dataShardCount + parityShardCount
	for i := 0; i < len(chunks); i += shardSize {

		count := 0
		for j := i; j < (i + shardSize); j++ {
			if chunks[j].Payload != nil {
				count++
			}
		}

		if count < dataShardCount {
			return false
		}

	}

	return true
}

func ReconstructMissingChunks(chunks []BlockChunk, dataShardCount int, parityShardCount int) []BlockChunk {

	totalShardCount := dataShardCount + parityShardCount

	var dataChunks []BlockChunk

	for i := 0; i < len(chunks); i += totalShardCount {

		data := make([][]byte, totalShardCount)

		enc, err := reedsolomon.New(dataShardCount, parityShardCount)
		if err != nil {
			panic(err)
		}

		c := chunks[i : i+totalShardCount]
		for j := range c {
			data[j] = c[j].Payload
		}

		// recunstructs missing chunks
		err = enc.Reconstruct(data)
		if err != nil {
			panic(err)
		}

		for j := 0; j < dataShardCount; j++ {
			// we do not need any other information
			dataChunks = append(dataChunks, BlockChunk{Payload: data[j]})
		}

	}

	return dataChunks
}

func addParityShards(chunks []BlockChunk, dataShardCount int, parityShardCount int) []BlockChunk {

	totalShardCount := dataShardCount + parityShardCount

	dataLength := len(chunks[0].Payload)
	data := make([][]byte, totalShardCount)
	for i := range data {
		if i < len(chunks) {
			data[i] = chunks[i].Payload
		} else {
			data[i] = make([]byte, dataLength)
		}
	}

	// creates encoder
	enc, err := reedsolomon.New(dataShardCount, parityShardCount)
	if err != nil {
		panic(err)
	}

	// creates parity shards
	err = enc.Encode(data)
	if err != nil {
		panic(err)
	}

	// creates parity shards
	for i := dataShardCount; i < totalShardCount; i++ {

		parityChunkData := data[i]
		parityChunk := BlockChunk{
			Payload:       parityChunkData,
			PayloadLength: len(parityChunkData),
		}

		chunks = append(chunks, parityChunk)
	}

	return chunks
}
