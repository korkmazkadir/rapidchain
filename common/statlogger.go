package common

import (
	"log"
	"time"
)

type StatLogger struct {
	round      int
	roundStart time.Time

	nodeID int
}

func NewStatLogger(nodeID int) *StatLogger {
	return &StatLogger{nodeID: nodeID}
}

func (s *StatLogger) NewRound(round int) {
	s.round = round
	s.roundStart = time.Now()
}

func (s *StatLogger) LogPropose(elapsedTime int64) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, s.round, "PROPOSE", elapsedTime)
}

func (s *StatLogger) LogBlockReceive(elapsedTime int64) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, s.round, "BLOCK_RECEIVED", elapsedTime)
}

func (s *StatLogger) LogEcho(elapsedTime int64) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, s.round, "ECHO", elapsedTime)
}

func (s *StatLogger) LogAccept(elapsedTime int64) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, s.round, "ACCEPT", elapsedTime)
}

func (s *StatLogger) LogEndOfRound() {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, s.round, "END_OF_ROUND", time.Since(s.roundStart).Milliseconds())
}
