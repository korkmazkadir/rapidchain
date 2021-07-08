package registery

import (
	"net/rpc"

	"github.com/korkmazkadir/rapidchain/common"
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

	registeryClient := RegistryClient{rpcClient: rpcClient, nodeInfo: currentNodeInfo}

	return registeryClient
}

// RegisterNode registers a node and returns assigned node ID
func (rc RegistryClient) RegisterNode() int {

	err := rc.rpcClient.Call("NodeRegistry.Register", rc.nodeInfo, &rc.nodeInfo)
	if err != nil {
		panic(err)
	}

	return rc.nodeInfo.ID
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

func (rc RegistryClient) UploadStats(statList common.StatList) {

	err := rc.rpcClient.Call("NodeRegistry.UploadStats", statList, nil)
	if err != nil {
		panic(err)
	}
}
