package registery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/korkmazkadir/rapidchain/common"
)

type StatKeeper struct {
	foderName string
}

func NewStatKeeper(config NodeConfig) *StatKeeper {

	// creates folder to save stats
	folderName := time.Now().Format("2006-01-02T15:04:05")
	err := os.Mkdir(folderName, 0755)
	if err != nil {
		panic(err)
	}

	statKeeper := &StatKeeper{foderName: folderName}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		panic(err)
	}

	// creates config file
	err = ioutil.WriteFile(statKeeper.GetConfigFilePath(), configJSON, 0644)
	if err != nil {
		panic(err)
	}

	return statKeeper
}

func (s *StatKeeper) SaveStats(statList common.StatList) {

	// writes node info to the filer
	nodeInfo := getNodeInfoString(statList.IPAddress, statList.PortNumber, statList.NodeID)

	nodeInfoFile, err := os.OpenFile(s.GetNodesFilePath(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}

	_, err = nodeInfoFile.WriteString(nodeInfo)
	if err != nil {
		panic(err)
	}
	nodeInfoFile.Close()
	if err != nil {
		panic(err)
	}

	// writes stats to the file
	statsFile, err := os.OpenFile(s.GetStatsFilePath(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}

	for _, e := range statList.Events {
		eventString := getEventString(statList.NodeID, e)
		_, err = statsFile.WriteString(eventString)
		if err != nil {
			panic(err)
		}
	}

	statsFile.Close()
	if err != nil {
		panic(err)
	}

}

func getNodeInfoString(ipAddress string, portNumber int, nodeID int) string {
	return fmt.Sprintf("%d\t%s\t%d\n", nodeID, ipAddress, portNumber)
}

func getEventString(nodeID int, event common.Event) string {
	return fmt.Sprintf("%d\t%d\t%s\t%d\n", nodeID, event.Round, event.Type, event.ElapsedTime)
}

func (s *StatKeeper) GetConfigFilePath() string {
	return fmt.Sprintf("./%s/config.json", s.foderName)
}

func (s *StatKeeper) GetStatsFilePath() string {
	return fmt.Sprintf("./%s/stats.log", s.foderName)
}

func (s *StatKeeper) GetNodesFilePath() string {
	return fmt.Sprintf("./%s/nodes.txt", s.foderName)
}
