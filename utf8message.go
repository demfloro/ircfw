package ircfw

import (
	"strings"
	"time"
)

type utf8message struct {
	prefix, cmd string
	params      []string
	client      *Client
}

func parseParams(line string) (result []string) {
	if len(line) == 0 {
		return
	}
	if line[0] == ':' {
		result = []string{strings.TrimSpace(line[1:])}
		return
	}
	i := strings.Index(line, " :")
	if i == -1 {
		result = strings.Split(strings.TrimSpace(line), " ")
		return
	}
	params_wo_spaces, last_param := line[:i], line[i:][2:]
	result = strings.Split(strings.TrimSpace(params_wo_spaces), " ")
	result = append(result, last_param)
	return
}

func parseUTF8Message(bytes []byte, client *Client) (msg message, err error) {
	line := string(bytes)
	if err = validate(line); err != nil {
		return
	}
	var (
		prefix, cmd string
		params      []string
	)
	if line[0] == ':' {
		prefix, line = pop(line[1:], " ")
		cmd, line = pop(line, " ")
		params = parseParams(line)
	} else {
		cmd, line = pop(line, " ")
		params = parseParams(line)
	}
	msg = utf8message{
		prefix: prefix,
		cmd:    cmd,
		params: params,
		client: client,
	}
	return
}

// Implemented this way to deny format string injections in the future
func (m utf8message) Export() []byte {
	var b strings.Builder
	b.Grow(MAXMSGSIZE)
	b.WriteString(m.cmd)
	if m.cmd == "PONG" || m.cmd == "PING" || m.cmd == "NICK" || m.cmd == "QUIT" {
		b.WriteString(" :")
		b.WriteString(strings.Join(m.params, " "))
		b.WriteString("\r\n")
		return []byte(b.String())
	}
	b.WriteString(" ")
	b.WriteString(strings.Join(m.params[:len(m.params)-1], " "))
	b.WriteString(" :")
	b.WriteString(m.params[len(m.params)-1])
	b.WriteString("\r\n")
	return []byte(b.String())
}

func (m utf8message) fetchChannel() *Channel {
	var chanName string
	switch m.cmd {
	case "332":
		chanName = m.params[1]
	case "353":
		chanName = m.params[2]
	default:
		chanName = m.params[0]
	}
	return m.client.fetchChannel(chanName)
}

func (m utf8message) Msg() Msg {
	channel := m.Channel()
	if channel == nil {
		channel = m.client.private
	}
	return ircMsg{
		time:    time.Now(),
		prefix:  m.Prefix(),
		text:    []string{strings.TrimSpace(m.params[1])},
		channel: channel,
		client:  m.client,
		utf8:    true,
	}
}

func newUTF8Message(cmd []byte, params [][]byte, client *Client) message {
	var uparams []string
	for _, param := range params {
		uparams = append(uparams, string(param))
	}
	return utf8message{
		cmd:    string(cmd),
		params: uparams,
		client: client,
	}
}

func (m utf8message) Cmd() string {
	return m.cmd
}

func (m utf8message) Nick() string {
	idx := strings.Index(m.prefix, "!")
	if idx == -1 {
		return m.prefix
	}
	return m.prefix[:idx]
}

func (m utf8message) MyNick() string {
	return m.client.Nick()
}

func (m utf8message) Channel() *Channel {
	m.client.Lock()
	defer m.client.Unlock()
	return m.fetchChannel()
}

func (m utf8message) Client() *Client {
	return m.client
}

func (m utf8message) Text() string {
	return m.params[1]
}

func (m utf8message) Params() []string {
	return m.params
}

func (m utf8message) Prefix() string {
	return m.prefix
}
