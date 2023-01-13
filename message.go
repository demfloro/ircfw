package ircfw

import (
	"fmt"
	"time"
)

func parseMessage(line []byte, deadline time.Time, client *Client) (msg message, err error) {
	if hasNULL(string(line)) {
		return nil, fmt.Errorf("contains NULL")
	}
	return parseUTF8Message(line, deadline, client)
}

func newMessage(cmd []byte, params [][]byte, deadline time.Time, client *Client) message {
	return newUTF8Message(cmd, params, deadline, client)
}
