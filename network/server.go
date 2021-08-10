package network

import (
	"github.com/korkmazkadir/rapidchain/common"
)

type P2PServer struct {
	demux *common.Demux
}

func NewServer(demux *common.Demux) *P2PServer {
	server := &P2PServer{demux: demux}
	return server
}

func (s *P2PServer) HandleBlockChunk(chunk *common.BlockChunk, reply *int) error {

	s.demux.EnqueBlockChunk(*chunk)

	return nil
}

func (s *P2PServer) HandleVote(vote *common.Vote, reply *int) error {

	s.demux.EnqueVote(*vote)

	return nil
}

func (s *P2PServer) HandleConnectRequest(connReq *common.ConnectRequest, reply *int) error {

	s.demux.EnqueConnectRequest(*connReq)

	return nil
}

func (s *P2PServer) HandleStateUpdate(stateUpdate *common.StateUpdate, reply *int) error {

	s.demux.EnqueStateUpdate(*stateUpdate)

	return nil
}
