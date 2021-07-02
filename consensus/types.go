package consensus

type Block struct {
	// Publick Key of the issuer
	Issuer []byte

	PrevBlockHash []byte

	Round int

	Payload []byte

	// Signature of the issuer
	Signature []byte
}

type BlockChunk struct {
	// Publick Key of the issuer
	Issuer []byte

	Round int

	ChunkIndex int

	Authenticator ChunkAuthenticator

	Payload []byte

	Signature []byte

	hash []byte
}

type Vote struct {

	// Publick Key of the issuer
	Issuer []byte

	Signature []byte

	hash []byte
}

type ChunkAuthenticator struct {
	// The last element is root
	Hashes [][]byte

	SignatureOnRoot []byte
}
