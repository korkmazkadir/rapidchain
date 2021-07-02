package consensus

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
)

func chunkBlock(block Block, numberOfChunks int) []BlockChunk {

	blockBytes := encodeToBytes(block)
	chunks := constructChunks(block, blockBytes, numberOfChunks)

	return chunks
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

		chunk := BlockChunk{
			Issuer:     block.Issuer,
			Round:      block.Round,
			ChunkIndex: i,
			Payload:    blockBytes[startIndex:endIndex],
			// TODO hash:
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
