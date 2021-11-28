package ircfw

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/tomb.v2"
)

func (c *Client) Debug(format string, params ...interface{}) {
	c.logger.Debug(format, params...)
}

func (c *Client) Logf(format string, params ...interface{}) {
	c.logger.Logf(format, params...)
}

func (c *Client) Wait() error {
	return c.tomb.Wait()
}

func (c *Client) Quit(reason string) {
	c.sendMessage("QUIT", []string{reason})
	c.Lock()
	c.killChannels()
	c.Unlock()
	c.private.kill()
	c.tomb.Kill(fmt.Errorf("user request: %q", reason))
}

func (c *Client) Join(ctx context.Context, chanName string) (*Channel, error) {
	err := validateChannel(chanName)
	if err != nil {
		return nil, fmt.Errorf("invalid channel name: %w", err)
	}
	if channel := c.fetchChannel(chanName); channel != nil && channel.isStarted() {
		return channel, nil
	}
	// Stall until initial message exchange with server finishes
	// without this client tries to join too early and server rejects it
	<-c.started
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
	ctx := c.tomb.Context(nil)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
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

func NewClient(opts ...Option) (*Client, context.CancelFunc) {
	conf := defaultConfig()
	for _, opt := range opts {
		opt(&conf)
	}
	t, _ := tomb.WithContext(conf.context)
	c := Client{
		tomb:         t,
		name:         conf.nick + "@" + conf.socket.RemoteAddr().String(),
		nickservPass: conf.nickservPass,
		socket:       conf.socket,
		logger:       conf.logger,
		channels:     make(map[string]*Channel),
		reads:        make(chan message, 32),
		writes:       make(chan message, 32),
		params:       make(map[string]string),
		charmap:      conf.charmap,
		handler:      conf.handler,
		started:      make(chan struct{}),
		aliveTimeout: 2 * time.Minute,
	}
	c.tomb.Go(c.serveLoop)
	c.tomb.Go(c.serveLoop)
	c.tomb.Go(c.writeLoop)
	c.tomb.Go(c.readLoop)
	c.tomb.Go(c.pingLoop)
	c.sendPass(conf.password)
	c.sendNick(conf.nick)
	c.sendUser(conf.ident, conf.realName)
	c.initPrivate()
	cancel := func() {
		t.Kill(fmt.Errorf("cancelled"))
		c.socket.Close()
	}
	return &c, cancel
}
