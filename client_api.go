package ircfw

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/text/encoding/charmap"
)

func (c *Client) Debug(format string, params ...interface{}) {
	c.logger.Debug(format, params...)
}

func (c *Client) Logf(format string, params ...interface{}) {
	c.logger.Logf(format, params...)
}

func (c *Client) Wait() error {
	c.wg.Wait()
	if c.err != nil {
		return c.err
	}
	return nil
}

func (c *Client) Quit(reason string) {
	c.sendMessage("QUIT", []string{reason})
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
		c.Debug("Attempt to set invalid nick: %q", nick)
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
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Time{}
	}
	select {
	case <-ctx.Done():
		return
	case c.writes <- newMessage([]byte(cmd), bparams, deadline, c):
		return
	}
}

func (c *Client) Whois(nick string) {
	if err := validateNick(nick); err != nil {
		c.Debug("Whoising %q: %w", nick, err)
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

func NewClient(ctx context.Context, nick string, ident string, realName string, password string, nickservPass string, socket net.Conn, logger Logger, handler MsgHandler, charmap *charmap.Charmap) (*Client, context.CancelFunc) {
	ctx, ctxcancel := context.WithCancel(ctx)
	cancel := func() {
		socket.Close()
		ctxcancel()
	}
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
		aliveTimeout: 2 * time.Minute,
	}
	c.wg.Add(5)
	go c.serveLoop(ctx)
	go c.serveLoop(ctx)
	go c.writeLoop(ctx)
	go c.readLoop(ctx)
	go c.pingLoop(ctx)
	c.sendPass(password)
	c.sendNick(nick)
	c.sendUser(ident, realName)
	c.initPrivate()
	return &c, cancel
}
