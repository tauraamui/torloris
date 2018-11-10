package main

import (
	"bufio"
	"os"

	"github.com/tacusci/logging"
)

var torEntryNodeIPs []string

func main() {
	file, err := os.Open("entrynodeslist.txt")
	if err != nil {
		logging.ErrorAndExit(err.Error())
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		torEntryNodeIPs = append(torEntryNodeIPs, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logging.ErrorAndExit(err.Error())
	}
}
