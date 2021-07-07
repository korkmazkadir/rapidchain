package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/korkmazkadir/rapidchain/registery"
)

func main() {

	// Read this from a file
	nodeConfig := registery.NodeConfig{
		NodeCount:       5,
		EpochSeed:       []byte{1, 2, 3, 4, 5},
		EndRound:        1000000,
		GossipFanout:    4,
		BlockSize:       2097152,
		BlockChunkCount: 128,
	}

	nodeRegistry := registery.NewNodeRegistry(nodeConfig)

	err := rpc.Register(nodeRegistry)
	if err != nil {
		panic(err)
	}

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	log.Printf("registery service started and listening on :1234\n")

	for {
		conn, _ := l.Accept()
		go func() {
			rpc.ServeConn(conn)
			//address := conn.RemoteAddr().String()
			//nodeRegistry.Unregister(address)
		}()
	}
}
