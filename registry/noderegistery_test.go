package registry

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRegistry(t *testing.T) {

	nodeConfig := NodeConfig{
		NodeCount:    10,
		EpochSeed:    []byte{1, 2, 3, 4, 5},
		EndRound:     10,
		GossipFanout: 16,
	}

	nodeRegistry := NewNodeRegistry(nodeConfig)

	// test register function
	nodeInfo := &NodeInfo{IPAddress: "abc", PortNumber: 6349}
	err := nodeRegistry.Register(nodeInfo, nodeInfo)
	if err != nil {
		t.Error(err)
	}

	if nodeInfo.ID == 0 {
		t.Error("registery did not assign a node id")
	}

	// test get config function
	retrievedConfig := &NodeConfig{}
	err = nodeRegistry.GetConfig(nodeInfo, retrievedConfig)
	if err != nil {
		t.Error(err)
	}

	if bytes.Equal(nodeConfig.Hash(), retrievedConfig.Hash()) == false {
		t.Errorf("could not retreive correct node config. expected: %v retreived: %v", nodeConfig, retrievedConfig)
	}

	// test get node list function
	nodeList := &NodeList{}
	err = nodeRegistry.GetNodeList(nodeInfo, nodeList)
	if err != nil {
		t.Error(err)
	}

	if len(nodeList.Nodes) != 1 {
		t.Errorf("ecpecting one node in the node list, received %d nodes", len(nodeList.Nodes))
	}

	receivedNode := nodeList.Nodes[0]
	if receivedNode.IPAddress != nodeInfo.IPAddress || receivedNode.PortNumber != nodeInfo.PortNumber {
		t.Errorf("node list is not correct; ecpected node %v, reveived node %v \n", nodeInfo, receivedNode)
	}

	nodeRegistry.Unregister(fmt.Sprintf("%s:%d", nodeInfo.IPAddress, nodeInfo.PortNumber))
	nodeList = &NodeList{}
	err = nodeRegistry.GetNodeList(nodeInfo, nodeList)
	if err != nil {
		t.Error(err)
	}

	if len(nodeList.Nodes) != 0 {
		t.Error("could not unregister the node", len(nodeList.Nodes))
	}

}
