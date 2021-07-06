package registry

import (
	"bytes"
	"log"
	"net"
	"net/rpc"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	// compy from cmd/registry/main.go
	nodeConfig := nodeConfigTestInstance()

	nodeRegistry := NewNodeRegistry(nodeConfig)

	rpc.Register(nodeRegistry)

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	log.Printf("registery service started and listening on :1234\n")

	go func() {
		for {
			conn, _ := l.Accept()
			go func() {
				rpc.ServeConn(conn)
				address := conn.RemoteAddr().String()
				nodeRegistry.Unregister(address)
			}()
		}
	}()

	os.Exit(m.Run())
}

func nodeConfigTestInstance() NodeConfig {
	nodeConfig := NodeConfig{
		NodeCount:    10,
		EpochSeed:    []byte{1, 2, 3, 4, 5},
		EndRound:     10,
		GossipFanout: 16,
	}
	return nodeConfig
}

func TestRegistryClient(t *testing.T) {

	// test register function
	nodeInfo := NodeInfo{IPAddress: "abc", PortNumber: 6349}
	registryAddress := "localhost:1234"
	registryClient := NewRegistryClient(registryAddress, nodeInfo)

	registryClient.RegisterNode()

	// test get node config
	nodeConfig := nodeConfigTestInstance()
	retrievedConfig := registryClient.GetConfig()

	if bytes.Equal(nodeConfig.Hash(), retrievedConfig.Hash()) == false {
		t.Errorf("could not retreive correct node config. expected: %v retreived: %v", nodeConfig, retrievedConfig)
	}

	// test get node list function
	nodes := registryClient.GetNodeList()

	if len(nodes) != 1 {
		t.Errorf("ecpecting one node in the node list, received %d nodes", len(nodes))
	}

	receivedNode := nodes[0]
	if receivedNode.IPAddress != nodeInfo.IPAddress || receivedNode.PortNumber != nodeInfo.PortNumber {
		t.Errorf("node list is not correct; ecpected node %v, reveived node %v \n", nodeInfo, receivedNode)
	}

}
