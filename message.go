package ircfw

import (
	"fmt"
)

// message factories
func parseMessage(line []byte, client *Client) (msg message, err error) {
	if hasNULL(string(line)) {
		return nil, fmt.Errorf("contains NULL")
	}
	if client.charmap != nil && !isUTF8(line) {
		return parseByteMessage(line, client)
	}
	return parseUTF8Message(line, client)
}

func newMessage(cmd []byte, params [][]byte, client *Client) message {
	if client.charmap == nil {
		return newUTF8Message(cmd, params, client)
	}
	for _, param := range params {
		if !isUTF8(param) {
			return newByteMessage(cmd, params, client)
		}
	}
	return newUTF8Message(cmd, params, client)
}
