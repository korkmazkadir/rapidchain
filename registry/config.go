package registry

type NodeConfig struct {
	NodeCount int

	EpochSeed []byte

	EndRound int

	GossipFanout int
}
