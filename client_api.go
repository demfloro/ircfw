package ircfw

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/text/encoding/charmap"
)

func (c *Client) Log(format string, params ...interface{}) {
	c.logger.Print(fmt.Errorf(format, params...))
}

func (c *Client) Wait() {
	c.wg.Wait()
}

func (c *Client) Quit(reason string) {
	c.sendMessage("QUIT", []string{reason})
	c.Lock()
	for name, channel := range c.channels {
		channel.kill()
		delete(c.channels, name)
	}
	c.Unlock()
	safeClose(c.started)
	c.private.kill()
	c.quit()
}

func (c *Client) Join(ctx context.Context, chanName string) (*Channel, error) {
	err := validateChannel(chanName)
	if err != nil {
		return nil, fmt.Errorf("Invalid channel name: %w", err)
	}
	if channel := c.fetchChannel(chanName); channel != nil && channel.isStarted() {
		return channel, nil
	}
	// Stall until initial message exchange with server finishes
	// without this client tries to join too early and server rejects it
	select {
	case <-c.started:
	}
	return c.joinChannel(ctx, chanName)
}

func (c *Client) extractNick() string {
	nick, _ := pop(c.prefix, "!")
	return nick
}

func (c *Client) Nick() string {
	c.Lock()
	defer c.Unlock()
	return c.extractNick()
}

func (c *Client) SetNick(nick string) {
	err := validateNick(nick)
	if err != nil {
		c.Log("Attempt to set invalid nick: %q", nick)
		return
	}
	c.sendNick(nick)
}

func (c *Client) sendMessage(cmd string, params []string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.sendMessageContext(ctx, cmd, params)
	cancel()
}

func (c *Client) sendMessageContext(ctx context.Context, cmd string, params []string) {
	var bparams [][]byte
	for _, param := range params {
		bparams = append(bparams, []byte(param))
	}
	select {
	case <-ctx.Done():
		return
	case c.writes <- newMessage([]byte(cmd), bparams, c):
		return
	}
}

func (c *Client) Whois(nick string) {
	if err := validateNick(nick); err != nil {
		c.Log("Whoising %q: %#v", nick, err)
		return
	}
	c.sendMessage("WHOIS", []string{nick})
}

func (c *Client) Prefix() string {
	c.Lock()
	defer c.Unlock()
	return c.prefix
}

func (c *Client) Motd() []string {
	return c.motd
}

func (c *Client) String() string {
	return c.name
}

func (c *Client) UpdateMode(target string, mode string) {
	c.Lock()
	if target == c.extractNick() {
		c.mode = mode
	}
	c.Unlock()
}

func NewClient(ctx context.Context, nick string, ident string, realName string, password string, nickservPass string, socket net.Conn, logger *log.Logger, handler MsgHandler, charmap *charmap.Charmap) (*Client, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	c := Client{
		name:         nick + "@" + socket.RemoteAddr().String(),
		nickservPass: nickservPass,
		socket:       socket,
		logger:       logger,
		channels:     make(map[string]*Channel),
		reads:        make(chan message, 32),
		writes:       make(chan message, 32),
		params:       make(map[string]string),
		charmap:      charmap,
		handler:      handler,
		started:      make(chan struct{}),
		quit:         cancel,
	}
	c.wg.Add(4)
	go c.serveLoop(ctx)
	go c.serveLoop(ctx)
	go c.writeLoop(ctx)
	go c.readLoop(ctx)
	c.sendPass(password)
	c.sendNick(nick)
	c.sendUser(ident, realName)
	c.initPrivate()
	return &c, cancel
}
