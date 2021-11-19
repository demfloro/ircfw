package ircfw

import (
	"bufio"
	"context"
	"errors"
	"time"

	"gopkg.in/tomb.v2"
)

var (
	ErrReadsClosed  = errors.New("c.reads closed")
	ErrWritesClosed = errors.New("c.writes closed")
	ErrTimeout      = errors.New("server timed out")
)

// Meant to run in separate goroutine
func (c *Client) writeLoop() error {
	var zero time.Time
	for {
		select {
		case <-c.tomb.Dying():
			c.Debug("writeLoop dying")
			return tomb.ErrDying
		case msg, open := <-c.writes:
			if !open {
				c.Debug("c.writes closed")
				c.socket.Close()
				return ErrWritesClosed
			}
			deadline := msg.Deadline()
			raw := msg.Export()
			c.Debug("writing raw: %q", string(raw))
			c.socket.SetWriteDeadline(deadline)
			_, err := c.socket.Write(raw)
			if err != nil {
				c.Debug("error writing to socket: %q", err)
				return err
			}
			c.socket.SetWriteDeadline(zero)
		}
	}
}

// Meant to run in separate goroutine
func (c *Client) readLoop() error {
	in := bufio.NewScanner(c.socket)
	in.Buffer(make([]byte, MAXMSGSIZE), MAXMSGSIZE)
	in.Split(scanMsg)
	for in.Scan() {
		line := in.Bytes()
		c.Debug("read raw: %q", string(line))
		t := time.Now()
		msg, err := parseMessage(line, t, c)
		if err != nil {
			c.Logf("Failed to parse: %q, err: %w", line, err)
			continue
		}
		select {
		case <-c.tomb.Dying():
			c.Debug("readLoop dying")
			c.socket.Close()
			return tomb.ErrDying
		case c.reads <- msg:
			c.Lock()
			c.lastMessage = t
			c.Unlock()
		}
	}
	return in.Err()
}

// Meant to run in separate goroutine
func (c *Client) pingLoop() error {
	pingFreq := c.aliveTimeout / 3
	ticker := time.NewTicker(pingFreq)
	for {
		select {
		case <-c.tomb.Dying():
			ticker.Stop()
			c.Debug("pingLoop dying")
			return tomb.ErrDying
		case now := <-ticker.C:
			c.Lock()
			if elapsed := now.Sub(c.lastMessage); elapsed >= c.aliveTimeout {
				c.Debug("Server timed out")
				c.Unlock()
				return ErrTimeout
			} else if elapsed >= pingFreq {
				c.ping([]string{c.extractNick()})
			}
			c.Unlock()
		}
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
		c.Debug("Duplicate attempt to create channel %q", chanName)
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
		return nil, ctx.Err()
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

func (c *Client) ping(params []string) {
	c.sendMessage("PING", params)
}

func (c *Client) pong(params []string) {
	c.sendMessage("PONG", params)
}

func (c *Client) killChannels() {
	for name, channel := range c.channels {
		channel.kill()
		delete(c.channels, name)
	}
}
