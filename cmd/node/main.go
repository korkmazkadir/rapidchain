package main

import (
	"encoding/base64"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
	"strings"
	"time"

	"github.com/korkmazkadir/rapidchain/common"
	"github.com/korkmazkadir/rapidchain/consensus"
	"github.com/korkmazkadir/rapidchain/network"
	"github.com/korkmazkadir/rapidchain/registery"
)

func main() {

	demux := common.NewDemultiplexer(0)
	server := network.NewServer(demux)

	err := rpc.Register(server)
	if err != nil {
		panic(err)
	}

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", "127.0.0.1:")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	// start serving
	go func() {
		for {
			conn, _ := l.Accept()
			go func() {
				rpc.ServeConn(conn)
			}()
		}
	}()

	log.Printf("p2p server started on %s\n", l.Addr().String())
	nodeInfo := getNodeInfo(l.Addr().String())

	registry := registery.NewRegistryClient("localhost:1234", nodeInfo)

	nodeInfo.ID = registry.RegisterNode()
	log.Printf("node registeration successful, assigned ID is %d\n", nodeInfo.ID)

	nodeConfig := registry.GetConfig()

	var nodeList []registery.NodeInfo

	for {
		nodeList = registry.GetNodeList()
		nodeCount := len(nodeList)
		if nodeCount == nodeConfig.NodeCount {
			break
		}
		time.Sleep(2 * time.Second)
		log.Printf("received node list %d/%d\n", nodeCount, nodeConfig.NodeCount)
	}

	peerSet := createPeerSet(nodeList, nodeConfig.GossipFanout, nodeInfo.ID)
	statLogger := common.NewStatLogger(nodeInfo.ID)
	rapidchain := consensus.NewRapidchain(demux, nodeConfig, peerSet, statLogger)

	runConsensus(rapidchain, nodeConfig.EndRound, nodeInfo.ID, nodeConfig.NodeCount, nodeConfig.BlockSize)

	// collects stats abd uploads to registry
	log.Printf("uploading stats to the registry\n")
	events := statLogger.GetEvents()
	statList := common.StatList{IPAddress: nodeInfo.IPAddress, PortNumber: nodeInfo.PortNumber, NodeID: nodeInfo.ID, Events: events}
	registry.UploadStats(statList)

	log.Printf("reached target round count. Shutting down in 1 minute\n")
	time.Sleep(1 * time.Minute)

	log.Printf("exiting as expected...\n")
}

func createPeerSet(nodeList []registery.NodeInfo, fanOut int, nodeID int) network.PeerSet {

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodeList), func(i, j int) { nodeList[i], nodeList[j] = nodeList[j], nodeList[i] })

	peerSet := network.PeerSet{}

	peerCount := 0
	for i := 0; i < len(nodeList); i++ {

		if peerCount == fanOut {
			break
		}

		peer := nodeList[i]
		if peer.ID == nodeID {
			continue
		}

		err := peerSet.AddPeer(peer.IPAddress, peer.PortNumber)
		if err != nil {
			panic(err)
		}
		log.Printf("new peer added: %s:%d ID %d\n", peer.IPAddress, peer.PortNumber, peer.ID)
		peerCount++
	}

	return peerSet
}

func getNodeInfo(netAddress string) registery.NodeInfo {
	tokens := strings.Split(netAddress, ":")

	ipAddress := tokens[0]
	portNumber, err := strconv.Atoi(tokens[1])
	if err != nil {
		panic(err)
	}

	return registery.NodeInfo{IPAddress: ipAddress, PortNumber: portNumber}
}

func runConsensus(rc *consensus.RapidchainConsensus, numberOfRounds int, nodeID int, nodeCount int, blockSize int) {

	time.Sleep(5 * time.Second)
	log.Println("Consensus staryed")

	// genesis block
	previousBlock := common.Block{Issuer: []byte("initial block"), Round: 0, Payload: []byte("hello world")}

	currentRound := 1
	for currentRound <= numberOfRounds {

		log.Printf("+++++++++ Round %d +++++++++++++++\n", currentRound)

		var block common.Block
		var err error

		proposerID := ((currentRound % nodeCount) + 1)

		if nodeID == proposerID {

			b := createBlock(currentRound, nodeID, previousBlock.Hash(), blockSize)

			block, err = rc.Propose(currentRound, b, previousBlock.Hash())
			if err != nil {
				panic(err)
			}

		} else {

			block, err = rc.Decide(currentRound, previousBlock.Hash())
			if err != nil {
				panic(err)
			}
		}

		previousBlock = block
		//log.Printf("decided block hash %x\n", encodeBase64(block.Hash()[:15]))

		currentRound++
		time.Sleep(2 * time.Second)
	}

}

// utils

func createBlock(round int, nodeID int, previousBlockHash []byte, blockSize int) common.Block {

	block := common.Block{
		Round:         round,
		Issuer:        []byte{byte(nodeID)},
		Payload:       getRandomByteSlice(blockSize),
		PrevBlockHash: previousBlockHash,
	}

	return block
}

func encodeBase64(hex []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(hex))
}

func getRandomByteSlice(size int) []byte {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return data
}
