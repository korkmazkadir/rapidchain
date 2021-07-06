package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/korkmazkadir/rapidchain/registry"
)

func main() {

	// Read this from a file
	nodeConfig := registry.NodeConfig{
		NodeCount:    10,
		EpochSeed:    []byte{1, 2, 3, 4, 5},
		EndRound:     10,
		GossipFanout: 16,
	}

	nodeRegistry := registry.NewNodeRegistry(nodeConfig)

	rpc.Register(nodeRegistry)

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
			address := conn.RemoteAddr().String()
			nodeRegistry.Unregister(address)
		}()
	}
}
