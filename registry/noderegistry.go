package registry

import (
	"log"
	"sync"
)

// NodeInfo keeps node info
type NodeInfo struct {
	IPAddress  string
	PortNumber int
}

type NodeList struct {
	Nodes []NodeInfo
}

type Logs struct {
	IPAddress  string
	PortNumber int
	Logs       []string
}

type Stats struct {
	IPAddress  string
	PortNumber int
	Stats      []string
}

type NodeRegistry struct {
	mutex           sync.Mutex
	registeredNodes []NodeInfo
}

// Register registers a node with specific node info
func (nr *NodeRegistry) Register(nodeInfo *NodeInfo, reply *int) error {

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	nr.registeredNodes = append(nr.registeredNodes, *nodeInfo)
	log.Printf("new node registered; ip address %s port number %d\n", nodeInfo.IPAddress, nodeInfo.PortNumber)

	return nil
}

// GetConfig is used to get config
func (nr *NodeRegistry) GetConfig(nodeInfo *NodeInfo, config *NodeConfig) error {

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	return nil
}

// GetNodeList returns node list
func (nr *NodeRegistry) GetNodeList(nodeInfo *NodeInfo, nodeList *NodeList) error {

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	nodeList.Nodes = append(nodeList.Nodes, nr.registeredNodes...)

	return nil
}

func (nr *NodeRegistry) UploadLogs(logs *Logs, reply *int) error {

	return nil
}

func (nr *NodeRegistry) UploadStats(stats *Stats, reply *int) error {

	return nil
}
