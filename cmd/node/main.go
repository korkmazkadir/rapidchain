package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/korkmazkadir/rapidchain/common"
	"github.com/korkmazkadir/rapidchain/consensus"
	"github.com/korkmazkadir/rapidchain/network"
	"github.com/korkmazkadir/rapidchain/registery"
)

func main() {

	hostname := getEnvWithDefault("NODE_HOSTNAME", "127.0.0.1")
	registryAddress := getEnvWithDefault("REGISTRY_ADDRESS", "localhost:1234")

	demux := common.NewDemultiplexer(0)
	server := network.NewServer(demux)

	err := rpc.Register(server)
	if err != nil {
		panic(err)
	}

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf("%s:", hostname))
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

	registry := registery.NewRegistryClient(registryAddress, nodeInfo)

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

	peerSet := createPeerSet(nodeList, nodeConfig.GossipFanout, nodeInfo)
	statLogger := common.NewStatLogger(nodeInfo.ID)
	rapidchain := consensus.NewRapidchain(demux, nodeConfig, peerSet, statLogger)

	runConsensus(rapidchain, nodeConfig.EndRound, nodeInfo.ID, nodeConfig.NodeCount, nodeConfig.LeaderCount, nodeConfig.BlockSize, nodeList)

	// collects stats abd uploads to registry
	log.Printf("uploading stats to the registry\n")
	events := statLogger.GetEvents()
	statList := common.StatList{IPAddress: nodeInfo.IPAddress, PortNumber: nodeInfo.PortNumber, NodeID: nodeInfo.ID, Events: events}
	registry.UploadStats(statList)

	log.Printf("reached target round count. Shutting down in 5 minute\n")
	time.Sleep(5 * time.Minute)

	log.Printf("exiting as expected...\n")
}

func createPeerSet(nodeList []registery.NodeInfo, fanOut int, nodeInfo registery.NodeInfo) network.PeerSet {

	var copyNodeList []registery.NodeInfo
	copyNodeList = append(copyNodeList, nodeList...)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(copyNodeList), func(i, j int) { copyNodeList[i], copyNodeList[j] = copyNodeList[j], copyNodeList[i] })

	peerSet := network.PeerSet{}

	peerCount := 0
	for i := 0; i < len(copyNodeList); i++ {

		if peerCount == fanOut {
			break
		}

		peer := copyNodeList[i]
		if peer.ID == nodeInfo.ID || peer.IPAddress == nodeInfo.IPAddress {
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

func runConsensus(rc *consensus.RapidchainConsensus, numberOfRounds int, nodeID int, nodeCount int, leaderCount int, blockSize int, nodeList []registery.NodeInfo) {

	time.Sleep(5 * time.Second)
	log.Println("Consensus started")

	// genesis block
	previousBlock := []common.Block{{Issuer: []byte("initial block"), Round: 0, Payload: []byte("hello world")}}

	currentRound := 1
	for currentRound <= numberOfRounds {

		log.Printf("+++++++++ Round %d +++++++++++++++\n", currentRound)

		var block []common.Block

		if isElectedAsLeader(nodeList, currentRound, nodeID, leaderCount) {
			log.Println("elected as leader")
			b := createBlock(currentRound, nodeID, hashBlock(previousBlock), blockSize, leaderCount)

			block = rc.Propose(currentRound, b, hashBlock(previousBlock))

		} else {

			block = rc.Decide(currentRound, hashBlock(previousBlock))

		}

		payloadSize := 0
		for i := range block {
			payloadSize += len(block[i].Payload)
		}

		log.Printf("appended payload size is %d bytes\n", payloadSize)

		previousBlock = block
		//log.Printf("decided block hash %x\n", encodeBase64(block.Hash()[:15]))

		currentRound++
		//time.Sleep(2 * time.Second)

		log.Printf("Appended block: %x\n", encodeBase64(hashBlock(block)[:15]))

	}

}

func hashBlock(blocks []common.Block) []byte {

	if len(blocks) == 1 {
		return blocks[0].Hash()
	}

	var hashSlice []byte
	for i := range blocks {
		hashSlice = append(hashSlice, blocks[i].Hash()...)
	}

	h := sha256.New()
	_, err := h.Write(hashSlice)
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

// utils

func createBlock(round int, nodeID int, previousBlockHash []byte, blockSize int, leaderCount int) common.Block {

	payloadSize := int(math.Ceil(float64(blockSize) / float64(leaderCount)))

	block := common.Block{
		Round:         round,
		Issuer:        []byte{byte(nodeID)},
		Payload:       getRandomByteSlice(payloadSize),
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

func getEnvWithDefault(key string, defaultValue string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		val = defaultValue
	}

	log.Printf("%s=%s\n", key, val)
	return val
}

func isElectedAsLeader(nodeList []registery.NodeInfo, round int, nodeID int, leaderCount int) bool {

	// assumes that node list is same for all nodes
	// shuffle the node list using round number as the source of randomness
	rand.Seed(int64(round))
	rand.Shuffle(len(nodeList), func(i, j int) { nodeList[i], nodeList[j] = nodeList[j], nodeList[i] })

	var electedLeaders []int
	for i := 0; i < leaderCount; i++ {
		electedLeaders = append(electedLeaders, nodeList[i].ID)
		if nodeList[i].ID == nodeID {
			log.Println("elected as leader")
			return true
		}
	}

	log.Printf("Elected leaders: %v\n", electedLeaders)

	return false
}
