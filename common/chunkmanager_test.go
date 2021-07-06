package common

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestChunkBlock(t *testing.T) {

	block := Block{
		Round:         0,
		Issuer:        getRandomByteSlice(32),
		Payload:       getRandomByteSlice(2097152),
		PrevBlockHash: getRandomByteSlice(32),
	}

	chunks := chunkBlock(block, 128)

	chunkCount := len(chunks)
	if chunkCount != 128 {
		t.Errorf("expected 128 chunk received %d chunk", chunkCount)
	}

	for _, c := range chunks {
		result, err := VerifyContentWithPath(c.Authenticator.MerkleRoot, c, c.Authenticator.Path, c.Authenticator.Index)
		if err != nil {
			t.Errorf("error occured during chunk verification chunk path %s", err)
		}

		if result == false {
			t.Errorf("failed to verify chunk expected TRUE, received %t Root %v", result, c.Authenticator.MerkleRoot)
		}

		//t.Logf("Len(payload) %d", len(c.Payload))
	}

	mergedBlock := mergeChunks(chunks)

	if mergedBlock.Round != block.Round {
		t.Errorf("expected round %d, received round %d", block.Round, mergedBlock.Round)
	}

	if bytes.Equal(mergedBlock.Issuer, block.Issuer) == false {
		t.Errorf("expected issuer %x, received issuer %x", block.Issuer, mergedBlock.Issuer)
	}

	if bytes.Equal(mergedBlock.Payload, block.Payload) == false {
		t.Errorf("payloads are not equal")
	}

	if bytes.Equal(mergedBlock.PrevBlockHash, block.PrevBlockHash) == false {
		t.Errorf("expected prev block hash %x, received  prev block hash %x", block.PrevBlockHash, mergedBlock.PrevBlockHash)
	}

	blockHash := block.Hash()
	mergedBlockHash := mergedBlock.Hash()

	if bytes.Equal(mergedBlockHash, blockHash) == false {
		t.Errorf("block hashes are not eques")
	}

}

func getRandomByteSlice(size int) []byte {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return data
}
