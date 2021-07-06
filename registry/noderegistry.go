package registry

import (
	"log"
	"strconv"
	"strings"
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
	// zipped content
	Logs []byte
}

type NodeRegistry struct {
	mutex           sync.Mutex
	registeredNodes []NodeInfo
	config          NodeConfig
}

func NewNodeRegistry(config NodeConfig) *NodeRegistry {

	return &NodeRegistry{config: config}
}

// Register registers a node with specific node info
func (nr *NodeRegistry) Register(nodeInfo *NodeInfo, reply *int) error {

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	nr.registeredNodes = append(nr.registeredNodes, *nodeInfo)
	log.Printf("new node registered; ip address %s port number %d, registered node count: %d\n", nodeInfo.IPAddress, nodeInfo.PortNumber, len(nr.registeredNodes))

	return nil
}

func (nr *NodeRegistry) Unregister(remoteAddress string) {
	addressParts := strings.Split(remoteAddress, ":")

	if len(addressParts) != 2 {
		log.Printf("unknown address format, node couldnot unregistered %s \n", remoteAddress)
		return
	}

	ipAddress := addressParts[0]
	portNumber, err := strconv.Atoi(addressParts[1])
	if err != nil {
		log.Printf("could not parse the port number, error: %s, portnumber: %s\n", err, addressParts[1])
		return
	}

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	nodeIndex := -1
	for i := range nr.registeredNodes {
		if nr.registeredNodes[i].IPAddress == ipAddress && nr.registeredNodes[i].PortNumber == portNumber {
			nodeIndex = i
			break
		}
	}

	if nodeIndex == -1 {
		log.Printf("could not find %s in the registered node list to unregister\n", remoteAddress)
		return
	}

	nr.registeredNodes = append(nr.registeredNodes[:nodeIndex], nr.registeredNodes[nodeIndex+1:]...)
	log.Printf("node %s unregistered successfully\n", remoteAddress)

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

	// writing byte to a file vs writing string
	// https://gobyexample.com/writing-files

	return nil
}
