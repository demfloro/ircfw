package ircfw

func (c *Channel) SetTopic(topic string) {
	if c.name == "" {
		c.Debug("Attempt to set topic on private")
		return
	}
	c.sendTopic(topic)
}

func (c *Channel) Topic() string {
	c.Lock()
	defer c.Unlock()
	return c.topic
}

// Calculate allowed message len for PRIVMSG
func (c *Channel) MsgLimit() int {
	var limit int
	// IRC message structure:
	// :prefix PRIVMSG ChannelName :text with spaces\r\n
	if c.name == "" {
		limit = MAXMSGSIZE - 1 - len(c.client.Prefix()) - 9 - 9 - 4
	} else {
		limit = MAXMSGSIZE - 1 - len(c.client.Prefix()) - 9 - len(c.name) - 4
	}
	if limit < 0 {
		return 0
	}
	return limit
}

func (c *Channel) Name() string {
	return c.name
}

func (c *Channel) Client() *Client {
	return c.client
}

func (c *Channel) Part() {
	if c.name == "" {
		return
	}
	c.client.sendMessage("PART", []string{c.name})
	close(c.quit)
}

func (c *Channel) Say(content string) {
	if content == "" {
		return
	}
	c.send <- NewIRCMsg([]string{content}, c, c.client, false)
}

func (c *Channel) Logf(format string, params ...interface{}) {
	c.client.Logf(format, params...)
}

func (c *Channel) Debug(format string, params ...interface{}) {
	c.client.Debug(format, params...)
}
