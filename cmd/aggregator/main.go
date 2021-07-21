package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/korkmazkadir/rapidchain/registery"
)

const configFile = "config.json"
const statFile = "stats.log"

func main() {

	globalStatFile := getGlobalStatFile()

	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Println(err)
			return err
		}

		if !info.IsDir() {
			return nil
		}

		config, err := getConfig(path)
		if err != nil {
			return nil
		}

		statFile, err := getStatFile(path)
		if err != nil {
			return nil
		}

		appendToLogs(config, statFile, globalStatFile)

		return nil
	})

	if err != nil {
		panic(err)
	}

	if err := globalStatFile.Close(); err != nil {
		panic(err)
	}

}

func appendToLogs(config registery.NodeConfig, stats *os.File, globalStatFile *os.File) {

	scanner := bufio.NewScanner(stats)
	prefix := fmt.Sprintf("%d\t%d\t%d\t", config.BlockSize, config.LeaderCount, config.BlockChunkCount)
	for scanner.Scan() {

		statLine := scanner.Text()
		globalStatLine := fmt.Sprintf("%s%s", prefix, statLine)
		_, err := fmt.Fprintln(globalStatFile, globalStatLine)
		if err != nil {
			panic(err)
		}
	}

}

func getGlobalStatFile() *os.File {

	file, err := os.OpenFile("experiment.stats", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}

	return file
}

func getStatFile(path string) (*os.File, error) {

	file, err := os.Open(fmt.Sprintf("%s/%s", path, statFile))
	return file, err
}

func getConfig(path string) (registery.NodeConfig, error) {

	config := registery.NodeConfig{}

	file, err := os.Open(fmt.Sprintf("%s/%s", path, configFile))
	if err != nil {
		return config, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
