package slowloris

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/tacusci/logging"
	"golang.org/x/net/html"
)

func NewClient() (*Client, error) {
	c := &Client{}
	if err := c.startup(); err != nil {
		return nil, err
	}
	return c, nil
}

type Client struct {
	tor        *tor.Tor
	HTTP       *http.Client
	dialCancel context.CancelFunc
}

func (c *Client) startup() error {
	t, err := tor.Start(nil, nil)
	if err != nil {
		return err
	}

	c.tor = t

	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	if err != nil {
		dialCancel()
		return err
	}

	c.dialCancel = dialCancel

	dialer, err := c.tor.Dialer(dialCtx, nil)
	if err != nil {
		dialCancel()
		return err
	}

	c.HTTP = &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}

	return nil
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
	if c.dialCancel != nil {
		c.dialCancel()
	}
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
