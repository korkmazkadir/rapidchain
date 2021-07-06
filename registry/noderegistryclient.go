package registry

import (
	"net/rpc"
)

type RegistryClient struct {
	rpcClient *rpc.Client
	nodeInfo  NodeInfo
}

func NewRegistryClient(registryAddress string, currentNodeInfo NodeInfo) RegistryClient {

	rpcClient, err := rpc.Dial("tcp", registryAddress)
	if err != nil {
		panic(err)
	}

	registryClient := RegistryClient{rpcClient: rpcClient, nodeInfo: currentNodeInfo}

	return registryClient
}

func (rc RegistryClient) RegisterNode() {

	err := rc.rpcClient.Call("NodeRegistry.Register", rc.nodeInfo, nil)
	if err != nil {
		panic(err)
	}

}

func (rc RegistryClient) GetConfig() NodeConfig {

	config := NodeConfig{}
	err := rc.rpcClient.Call("NodeRegistry.GetConfig", rc.nodeInfo, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func (rc RegistryClient) GetNodeList() []NodeInfo {

	nodeList := NodeList{}
	err := rc.rpcClient.Call("NodeRegistry.GetNodeList", rc.nodeInfo, &nodeList)
	if err != nil {
		panic(err)
	}

	return nodeList.Nodes
}
