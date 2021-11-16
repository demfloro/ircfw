package ircfw

import (
	"bytes"
	"time"
)

type bytemessage struct {
	prefix, cmd []byte
	params      [][]byte
	deadline    time.Time
	client      *Client
}

func (m bytemessage) Deadline() time.Time {
	return m.deadline
}

func parseByteParams(line []byte) (result [][]byte) {
	if len(line) == 0 {
		return
	}
	// 58 is ascii ':'
	if line[0] == 58 {
		result = [][]byte{bytes.TrimSpace(line[1:])}
		return
	}
	i := bytes.Index(line, []byte(":"))
	if i == -1 {
		result = bytes.Split(bytes.TrimSpace(line), []byte(" "))
		return
	}
	params_wo_spaces, last_param := line[:i], line[i:][1:]
	result = bytes.Split(bytes.TrimSpace(params_wo_spaces), []byte(" "))
	result = append(result, last_param)
	return
}

func parseByteMessage(line []byte, deadline time.Time, client *Client) (msg message, err error) {
	if err = validate(decode(line, client.charmap)); err != nil {
		return
	}
	var (
		prefix, cmd []byte
		params      [][]byte
	)
	// 58 is ascii ':'
	if line[0] == 58 {
		prefix, line = bytepop(line[1:], []byte(" "))
		cmd, line = bytepop(line, []byte(" "))
		params = parseByteParams(line)
	}
	msg = bytemessage{
		prefix:   prefix,
		cmd:      cmd,
		params:   params,
		deadline: deadline,
		client:   client,
	}
	return
}

// Implemented this way to deny format string injections in the future
func (m bytemessage) Export() []byte {
	b := make([]byte, 0, MAXMSGSIZE)
	b = append(b, m.cmd...)
	if bytes.Equal(m.cmd, []byte("PONG")) || bytes.Equal(m.cmd, []byte("PING")) ||
		bytes.Equal(m.cmd, []byte("NICK")) || bytes.Equal(m.cmd, []byte("QUIT")) {
		b = append(b, []byte(" :")...)
		b = append(b, bytes.Join(m.params, []byte(" "))...)
		b = append(b, []byte("\r\n")...)
		return b
	}
	b = append(b, []byte(" ")...)
	b = append(b, bytes.Join(m.params[:len(m.params)-1], []byte(" "))...)
	b = append(b, []byte(" :")...)
	b = append(b, m.params[len(m.params)-1]...)
	b = append(b, []byte("\r\n")...)
	return b
}

func (m bytemessage) fetchChannel() *Channel {
	var chanName []byte
	if bytes.Equal(m.cmd, []byte("332")) {
		chanName = m.params[1]
	} else if bytes.Equal(m.cmd, []byte("353")) {
		chanName = m.params[2]
	} else {
		chanName = m.params[0]
	}
	decodedChanName := decode(chanName, m.client.charmap)
	return m.client.fetchChannel(decodedChanName)
}

func (m bytemessage) Msg() Msg {
	channel := m.Channel()
	if channel == nil {
		channel = m.client.private
	}
	return ircMsg{
		time:     time.Now(),
		deadline: m.deadline,
		prefix:   m.Prefix(),
		text:     []string{decode(bytes.TrimSpace(m.params[1]), m.client.charmap)},
		channel:  channel,
		client:   m.client,
		utf8:     false,
	}
}

func newByteMessage(cmd []byte, params [][]byte, deadline time.Time, client *Client) message {
	return bytemessage{
		cmd:      cmd,
		params:   params,
		deadline: deadline,
		client:   client,
	}
}

func (m bytemessage) Cmd() string {
	return string(m.cmd)
}

func (m bytemessage) Nick() string {
	idx := bytes.Index(m.prefix, []byte("!"))
	if idx == -1 {
		return decode(m.prefix, m.Client().charmap)
	}
	return decode(m.prefix[:idx], m.Client().charmap)
}

func (m bytemessage) MyNick() string {
	return m.client.Nick()
}

func (m bytemessage) Channel() *Channel {
	m.client.Lock()
	defer m.client.Unlock()
	return m.fetchChannel()
}

func (m bytemessage) Client() *Client {
	return m.client
}

func (m bytemessage) Text() string {
	return decode(m.params[1], m.Client().charmap)
}

func (m bytemessage) Params() []string {
	var out []string
	for _, slice := range m.params {
		out = append(out, decode(slice, m.Client().charmap))
	}
	return out
}

func (m bytemessage) Prefix() string {
	return decode(m.prefix, m.Client().charmap)
}
