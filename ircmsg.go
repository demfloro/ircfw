package ircfw

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ircMsg struct {
	time     time.Time
	deadline time.Time
	prefix   string
	text     []string
	channel  *Channel
	client   *Client
	utf8     bool
}

func NewIRCMsg(text []string, channel *Channel, client *Client, utf8 bool, deadline time.Time) Msg {
	return ircMsg{
		time:     time.Now(),
		deadline: time.Time{},
		text:     text,
		channel:  channel,
		client:   client,
		utf8:     utf8,
	}
}

func (m ircMsg) Time() time.Time {
	return m.time
}

func (m ircMsg) Nick() string {
	return strings.Split(m.prefix, "!")[0]
}

func (m ircMsg) Prefix() string {
	return m.prefix
}

func (m ircMsg) Channel() *Channel {
	return m.channel
}

func (m ircMsg) Client() *Client {
	return m.client
}

func (m ircMsg) WrappedText() []string {
	lenLimit := m.Channel().MsgLimit()
	textLength := textLen(m.Text())
	if textLength <= lenLimit {
		return m.Text()
	}
	result := make([]string, 0, len(m.Text()))
	for _, line := range m.Text() {
		result = append(result, splitByLen(line, lenLimit, 0)...)
	}
	return result
}

func (m ircMsg) Text() []string {
	return m.text
}

func (m ircMsg) Logf(format string, params ...interface{}) {
	m.channel.Logf(format, params...)
}

func (m ircMsg) Debug(format string, params ...interface{}) {
	m.channel.Debug(format, params...)
}

func (m ircMsg) IsPrivate() bool {
	return m.channel.name == ""
}

func (m ircMsg) Reply(ctx context.Context, text []string) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Time{}
	}
	msg := ircMsg{
		time:     time.Now(),
		deadline: deadline,
		prefix:   m.Prefix(),
		text:     text,
		channel:  m.channel,
		client:   m.client,
		utf8:     m.utf8,
	}
	select {
	case <-ctx.Done():
		m.Logf("reply timed out: %#v", msg)
		return
	case m.channel.send <- msg:
	}
}

func (m ircMsg) String() string {
	return fmt.Sprintf("ircfw.ircMsg{time: %q, prefix: %q, channel: %q, client: %q, utf8: %v, text %q}", m.time.Format("2006-01-02 15:04:05"), m.prefix, m.channel.name, m.client.name, m.utf8, m.text)
}

func (m ircMsg) Messages() (messages []message) {
	chanName := m.channel.Name()
	if m.IsPrivate() {
		chanName = m.Nick()
	}
	if !m.utf8 || m.client.charmap != nil {
		for _, line := range m.WrappedText() {
			messages = append(messages, newMessage([]byte("PRIVMSG"), [][]byte{[]byte(chanName), encode(line, m.client.charmap)}, m.deadline, m.client))
		}
	} else {
		for _, line := range m.WrappedText() {
			messages = append(messages, newMessage([]byte("PRIVMSG"), [][]byte{[]byte(chanName), []byte(line)}, m.deadline, m.client))
		}
	}
	return
}
