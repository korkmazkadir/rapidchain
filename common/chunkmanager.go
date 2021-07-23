package common

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
)

func ChunkBlock(block Block, numberOfChunks int) []BlockChunk {

	blockBytes := encodeToBytes(block)
	chunks := constructChunks(block, blockBytes, numberOfChunks)
	return chunks
}

// mergeChunks assumes that sanity checks are done before calling this function
func MergeChunks(chunks []BlockChunk) Block {

	var blockData []byte
	for i := 0; i < len(chunks); i++ {
		blockData = append(blockData, chunks[i].Payload...)
	}

	return decodeToBlock(blockData)
}

func constructChunks(block Block, blockBytes []byte, numberOfChunks int) []BlockChunk {

	var chunks []BlockChunk
	chunkSize := int(math.Ceil(float64(len(blockBytes)) / float64(numberOfChunks)))

	if chunkSize == 0 {
		panic(fmt.Errorf("chunk payload size is 0"))
	}

	for i := 0; i < numberOfChunks; i++ {

		startIndex := i * chunkSize
		endIndex := startIndex + chunkSize

		var payload []byte
		if i < (numberOfChunks - 1) {
			payload = blockBytes[startIndex:endIndex]
		} else {
			payload = blockBytes[startIndex:]
		}

		chunk := BlockChunk{
			Round:      block.Round,
			ChunkCount: numberOfChunks,
			ChunkIndex: i,
			Payload:    payload,
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

// https://gist.github.com/SteveBate/042960baa7a4795c3565
func encodeToBytes(p interface{}) []byte {

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func decodeToBlock(data []byte) Block {

	block := Block{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&block)
	if err != nil {
		panic(err)
	}
	return block
}
