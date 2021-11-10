package ircfw

import (
	"bufio"
	"context"
	"fmt"
)

// Meant to run in separate goroutine
func (c *Client) writeLoop(ctx context.Context) {
	defer c.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, open := <-c.writes:
			if !open {
				c.Log("c.writes closed")
				return
			}
			raw := msg.Export()
			c.socket.Write(raw)
		}
	}
}

// Meant to run in separate goroutine
func (c *Client) readLoop(ctx context.Context) {
	defer c.wg.Done()
	in := bufio.NewScanner(c.socket)
	in.Buffer(make([]byte, MAXMSGSIZE), MAXMSGSIZE)
	in.Split(scanMsg)
	for in.Scan() {
		line := in.Bytes()
		msg, err := parseMessage(line, c)
		if err != nil {
			c.Log("Failed to parse: %v, err: %w", line, err)
			continue
		}
		select {
		case <-ctx.Done():
			return
		case c.reads <- msg:
		}
	}
	if err := in.Err(); err != nil {
		c.Log("Error in readLoop: %w", err)
		return
	}
}

func (c *Client) sendPass(password string) {
	if password == "" {
		return
	}
	c.sendMessage("PASS", []string{password})
}

func (c *Client) sendNick(nick string) {
	if nick == "" {
		return
	}
	c.sendMessage("NICK", []string{nick})
}

func (c *Client) sendUser(ident string, realName string) {
	c.sendMessage("USER", []string{ident, "0.0.0.0", "0.0.0.0", realName})
}

func (c *Client) fetchChannel(chanName string) *Channel {
	channel, ok := c.channels[chanName]
	if !ok {
		return nil
	}
	return channel
}

func (c *Client) createChannel(chanName string) *Channel {
	c.Lock()
	defer c.Unlock()
	if channel := c.fetchChannel(chanName); channel != nil {
		c.Log("Duplicate attempt to create channel %q", chanName)
		return channel
	}
	channel := newChannel(chanName, c)
	c.channels[chanName] = channel
	return channel

}

func (c *Client) joinChannel(ctx context.Context, chanName string) (channel *Channel, err error) {
	channel = c.createChannel(chanName)
	c.sendMessage("JOIN", []string{chanName})

	select {
	case <-ctx.Done():
		c.Lock()
		delete(c.channels, chanName)
		channel.kill()
		c.Unlock()
		return nil, fmt.Errorf("Timed out to join %q", chanName)
	case <-channel.quit:
		return nil, channel.err
	case <-channel.started:
		return channel, nil
	}
}

func (c *Client) setNick(nick string) {
	c.Lock()
	_, identHostname := pop(c.prefix, "!")
	c.prefix = join([]string{nick, identHostname}, "!")
	c.Unlock()
}

func (c *Client) setHostname(hostname string) {
	c.Lock()
	nick, identHostname := pop(c.prefix, "!")
	ident, _ := pop(identHostname, "@")
	c.prefix = nick + "!" + ident + "@" + hostname
	c.Unlock()
}

func (c *Client) initPrivate() {
	c.private = newChannel("", c)
	c.private.start()
}

func (c *Client) pong(params []string) {
	c.sendMessage("PONG", params)
}
