package consensus

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"

	"github.com/cbergoon/merkletree"
)

func chunkBlock(block Block, numberOfChunks int) []BlockChunk {

	blockBytes := encodeToBytes(block)
	chunks := constructChunks(block, blockBytes, numberOfChunks)
	createAuthenticators(chunks)

	return chunks
}

// mergeChunks assumes that sanity checks are done before calling this function
func mergeChunks(chunks []BlockChunk) Block {

	var blockData []byte
	for i := 0; i < len(chunks); i++ {
		blockData = append(blockData, chunks[i].Payload...)
	}

	return decodeToBlock(blockData)
}

func createAuthenticators(chunks []BlockChunk) {

	// construct merkletree
	var content []merkletree.Content
	for _, c := range chunks {
		content = append(content, c)
	}

	tree, err := merkletree.NewTree(content)
	if err != nil {
		panic(err)
	}

	// calculates the root of the merkle tree
	merkleRoot := tree.MerkleRoot()

	// creates merklepath for each chunk
	for i := 0; i < len(chunks); i++ {
		path, index, err := tree.GetMerklePath(chunks[i])
		if err != nil {
			panic(err)
		}

		authenticator := ChunkAuthenticator{MerkleRoot: merkleRoot, Path: path, Index: index}
		chunks[i].Authenticator = authenticator
	}

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
			Issuer:     block.Issuer,
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
