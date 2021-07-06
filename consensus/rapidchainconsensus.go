package consensus

import (
	"github.com/korkmazkadir/rapidchain/common"
	"github.com/korkmazkadir/rapidchain/network"
	"github.com/korkmazkadir/rapidchain/registry"
)

type RapidchainConsensus struct {
	demultiplexer *common.Demux
	nodeConfig    registry.NodeConfig
	peerSet       network.PeerSet
}
