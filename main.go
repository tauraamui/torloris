package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/tacusci/logging"
	"golang.org/x/net/html"
)

type torClient struct {
	t          *tor.Tor
	client     *http.Client
	dialCancel context.CancelFunc
}

func (tc *torClient) Init() error {
	t, err := tor.Start(nil, nil)
	if err != nil {
		return err
	}

	tc.t = t

	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	if err != nil {
		dialCancel()
		return err
	}

	tc.dialCancel = dialCancel

	dialer, err := tc.t.Dialer(dialCtx, nil)
	if err != nil {
		dialCancel()
		return err
	}

	tc.client = &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}

	return nil
}

func main() {
	torClient := torClient{}
	torClient.Init()

	resp, err := torClient.client.Get("https://check.torproject.org")
	if err != nil {
		logging.Error(fmt.Sprintf("%v", err))
	}
	defer resp.Body.Close()
	// Grab the <title>
	parsed, err := html.Parse(resp.Body)
	if err != nil {
		logging.Error(fmt.Sprintf("%v", err))
	}
	fmt.Printf("Title: %v\n", getTitle(parsed))

	torClient.dialCancel()
}

func run() error {
	// Start tor with default config (can set start conf's DebugWriter to os.Stdout for debug logs)
	fmt.Println("Starting tor and fetching title of https://check.torproject.org, please wait a few seconds...")
	t, err := tor.Start(nil, nil)
	if err != nil {
		return err
	}
	defer t.Close()
	// Wait at most a minute to start network and get
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dialCancel()
	// Make connection
	dialer, err := t.Dialer(dialCtx, nil)
	if err != nil {
		return err
	}
	httpClient := &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}
	// Get /
	resp, err := httpClient.Get("https://check.torproject.org")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Grab the <title>
	parsed, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Title: %v\n", getTitle(parsed))
	return nil
}

func getTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		var title bytes.Buffer
		if err := html.Render(&title, n.FirstChild); err != nil {
			panic(err)
		}
		return strings.TrimSpace(title.String())
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := getTitle(c); title != "" {
			return title
		}
	}
	return ""
}
