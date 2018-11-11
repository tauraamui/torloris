package main

import (
	"github.com/tacusci/logging"
	"github.com/tauraamui/torloris/slowloris"
)

func main() {
	client, err := slowloris.NewClient()

	if err != nil {
		logging.ErrorAndExit(err.Error())
	}

	logging.InfoNnlNoColor("Checking connected via Tor service... ")
	if client.CheckTorConnection() {
		logging.GreenOutput("Connected!\n")
	} else {
		logging.RedOutput("Not connected!\n")
	}
	client.Close()
}
