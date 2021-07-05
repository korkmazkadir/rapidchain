package consensus

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/cbergoon/merkletree"
)

const (
	// ProposeTag show a vote belogs to propose phase of the consensus instance
	ProposeTag = 'P'

	// EchoTag show a vote belogs to echo phase of the consensus instance
	EchoTag = 'E'

	// AcceptTag show a vote belogs to accept phase of the consensus instance
	AcceptTag = 'A'
)

// Block defines blockchain block structure
type Block struct {
	Issuer []byte

	PrevBlockHash []byte

	Round int

	Payload []byte
}

// Hash produces the digest of a Block.
// It considers all fields of a Block.
func (b *Block) Hash() []byte {

	str := fmt.Sprintf("%x,%x,%d,%x", b.Issuer, b.PrevBlockHash, b.Round, b.Payload)
	h := sha256.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

// AcceptProof proof of the accept. Should contain mf+1 echo messahes from different nodes for the
// same Merkleroot
type AcceptProof struct {
	EchoVotes []Vote
}

// Hash hashes a AcceptProof
func (ap AcceptProof) Hash() []byte {

	if len(ap.EchoVotes) == 0 {
		return nil
	}

	var hashSlice []byte
	for i := range ap.EchoVotes {
		hashSlice = append(hashSlice, ap.EchoVotes[i].Hash()...)
	}

	h := sha256.New()
	_, err := h.Write(hashSlice)
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

// Vote defines a consensus vote.
type Vote struct {

	// Publick Key of the issuer
	Issuer []byte

	Tag byte

	Round int

	BlockHash []byte

	Proof AcceptProof

	Signature []byte
}

// Hash hashes a vote
func (v Vote) Hash() []byte {

	str := fmt.Sprintf("%x,%d,%d,%x,%x", v.Issuer, v.Tag, v.Round, v.BlockHash, v.Proof.Hash())
	h := sha256.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

// BlockChunk defines a chunk of a block.
// BlockChunks disseminate fater in the gossip network because they are very small compared to a Block
type BlockChunk struct {
	// Publick Key of the issuer
	Issuer []byte

	// Round of the block
	Round int

	// The number of expected chunks to reconstruct a block
	ChunkCount int

	// Chunk index
	ChunkIndex int

	// Chunk authenticator to validate chunk
	Authenticator ChunkAuthenticator

	// Chunk payload
	Payload []byte

	// Signature on the hash of the BlockChunk
	Signature []byte

	payloadHash []byte
}

// Hash produces the digest of a BlockChunk.
// It considers all fields of a BlockChunk.
func (c BlockChunk) Hash() []byte {

	str := fmt.Sprintf("%x,%d,%d,%d,%x,%x", c.Issuer, c.Round, c.ChunkCount, c.ChunkIndex, c.Authenticator.Hash(), c.Payload)
	h := sha256.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

// CalculateHash is defined in merkletree interface.
// This method calculates the hash of the payload.
func (c BlockChunk) CalculateHash() ([]byte, error) {

	if c.payloadHash == nil {
		h := sha256.New()
		_, err := h.Write(c.Payload)
		if err != nil {
			return nil, err
		}

		c.payloadHash = h.Sum(nil)
	}

	return c.payloadHash, nil
}

// Equals is defined in merkletree interface
// This method compares payloads.
func (c BlockChunk) Equals(other merkletree.Content) (bool, error) {

	currentHash, err := c.CalculateHash()
	if err != nil {
		return false, nil
	}

	otherHash, err := other.CalculateHash()
	if err != nil {
		return false, nil
	}

	return bytes.Equal(currentHash, otherHash), nil
}

// ChunkAuthenticator contains all the information to authenticate a BlockChunk
type ChunkAuthenticator struct {

	// Root of the Merkle tree
	MerkleRoot []byte

	// Merkle path of the current chunk
	Path [][]byte

	// Path index
	Index []int64
}

// Hash produces the digest of a ChunkAuthenticator.
// It considers all fields of a ChunkAuthenticator.
func (c *ChunkAuthenticator) Hash() []byte {

	str := fmt.Sprintf("%x,%x,%v", c.MerkleRoot, c.Path, c.Index)
	h := sha256.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}
