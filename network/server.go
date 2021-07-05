package network

import (
	"github.com/korkmazkadir/rapidchain/consensus"
)

type P2PServer struct {
	demux *consensus.Demux
}

func NewServer(demux *consensus.Demux) *P2PServer {
	server := &P2PServer{demux: demux}
	return server
}

func (s *P2PServer) HandleBlockChunk(chunk *consensus.BlockChunk, reply *int) error {

	s.demux.EnqueBlockChunk(*chunk)

	return nil
}

func (s *P2PServer) HandleVote(vote *consensus.Vote, reply *int) error {

	s.demux.EnqueVote(*vote)

	return nil
}
