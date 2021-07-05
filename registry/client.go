package registry

import (
	"fmt"
	"net/rpc"
)

type RegistryClient struct {
	ipAddress  string
	portNumber int
	rpcClient  *rpc.Client
}

func NewRegistryClient(ipAddress string, portNumber int) (*RegistryClient, error) {

	rpcClient, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", ipAddress, portNumber))
	if err != nil {
		return nil, err
	}

	registryClient := &RegistryClient{ipAddress: ipAddress, portNumber: portNumber, rpcClient: rpcClient}

	return registryClient, nil
}

func (rc *RegistryClient) RegisterNode(ipAddress string, portNumber int) error {

	nodeInfo := NodeInfo{IPAddress: ipAddress, PortNumber: portNumber}
	return rc.rpcClient.Call("NodeRegistry.Register", nodeInfo, nil)
}
