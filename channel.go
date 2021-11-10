package ircfw

func newChannel(name string, client *Client) *Channel {
	c := &Channel{
		name:    name,
		names:   NewSet(),
		client:  client,
		send:    make(chan Msg, 8),
		receive: make(chan Msg, 8),
		started: make(chan struct{}),
		quit:    make(chan struct{}),
	}
	return c
}

func (c *Channel) start() {
	go c.rxLoop()
	go c.txLoop()
	close(c.started)
}

func (c *Channel) isStarted() bool {
	select {
	case <-c.started:
		return true
	default:
		return false
	}
}

func (c *Channel) setTopic(topic string) {
	c.topic = topic
}

func (c *Channel) sendTopic(topic string) {
	c.client.sendMessage("TOPIC", []string{c.name, topic})
}

func (c *Channel) queryTopic() {
	c.client.sendMessage("TOPIC", []string{c.name})
}

func (c *Channel) kill() {
	safeClose(c.quit)
}

func (c *Channel) String() string {
	return c.name
}

// meant to run in separate goroutine
func (c *Channel) rxLoop() {
	for {
		select {
		case <-c.quit:
			return
		case msg, open := <-c.receive:
			if !open {
				safeClose(c.quit)
				return
			}
			select {
			case <-c.quit:
				return
			default:
				c.client.handler(msg)
			}
		}
	}
}

// meant to run in separate goroutine
func (c *Channel) txLoop() {
	for {
		select {
		case <-c.quit:
			return
		case msg, open := <-c.send:
			if !open {
				c.Client().Log("%q send closed, killing loop", c.name)
				safeClose(c.quit)
				return
			}
			for _, message := range msg.Messages() {
				select {
				case <-c.quit:
					return
				case c.Client().writes <- message:
					continue

				}
			}
		}
	}

}
