package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math"

	"github.com/cbergoon/merkletree"
)

func ChunkBlock(block Block, numberOfChunks int) ([]BlockChunk, []byte) {

	blockBytes := encodeToBytes(block)
	chunks := constructChunks(block, blockBytes, numberOfChunks)
	merkleRootHash := createAuthenticators(chunks)

	return chunks, merkleRootHash
}

// mergeChunks assumes that sanity checks are done before calling this function
func MergeChunks(chunks []BlockChunk) Block {

	var blockData []byte
	for i := 0; i < len(chunks); i++ {
		blockData = append(blockData, chunks[i].Payload...)
	}

	return decodeToBlock(blockData)
}

// createAuthenticators returns mekle root
func createAuthenticators(chunks []BlockChunk) []byte {

	// construct merkletree
	var content []merkletree.Content
	for i := range chunks {
		content = append(content, chunks[i])
	}

	tree, err := merkletree.NewTree(content)
	if err != nil {
		panic(err)
	}

	// calculates the root of the merkle tree
	merkleRoot := tree.MerkleRoot()

	// creates merklepath for each chunk
	for i := range chunks {
		path, index, err := tree.GetMerklePath(chunks[i])
		if err != nil {
			panic(err)
		}

		authenticator := ChunkAuthenticator{MerkleRoot: merkleRoot, Path: path, Index: index}
		chunks[i].Authenticator = authenticator
	}

	return merkleRoot
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

// VerifyContentWithPath verifies content using path information comming from GetMerklePath function, and Merkle root.
func VerifyContentWithPath(merkleRoot []byte, content merkletree.Content, path [][]byte, index []int64) (bool, error) {

	if len(path) != len(index) {
		return false, fmt.Errorf("path or index argument is wrong")
	}

	calculatedRoot, err := content.CalculateHash()
	if err != nil {
		return false, err
	}

	// assumes this is same with merkletree package
	hashStrategy := sha256.New

	for i := 0; i < len(path); i++ {

		h := hashStrategy()
		if index[i] == 0 {
			_, err = h.Write(append(path[i], calculatedRoot...))
			calculatedRoot = h.Sum(nil)
		} else {
			_, err = h.Write(append(calculatedRoot, path[i]...))
			calculatedRoot = h.Sum(nil)
		}

		if err != nil {
			return false, err
		}

	}

	return bytes.Equal(merkleRoot, calculatedRoot), nil
}
