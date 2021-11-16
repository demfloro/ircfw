package ircfw

import (
	"fmt"
	"time"
)

// message factories
func parseMessage(line []byte, deadline time.Time, client *Client) (msg message, err error) {
	if hasNULL(string(line)) {
		return nil, fmt.Errorf("contains NULL")
	}
	if client.charmap != nil && !isUTF8(line) {
		return parseByteMessage(line, deadline, client)
	}
	return parseUTF8Message(line, deadline, client)
}

func newMessage(cmd []byte, params [][]byte, deadline time.Time, client *Client) message {
	if client.charmap == nil {
		return newUTF8Message(cmd, params, deadline, client)
	}
	for _, param := range params {
		if !isUTF8(param) {
			return newByteMessage(cmd, params, deadline, client)
		}
	}
	return newUTF8Message(cmd, params, deadline, client)
}
