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

	client.CheckTorConnection()
	client.Close()
}
