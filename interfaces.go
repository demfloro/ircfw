package ircfw

import (
	"context"
	"net"
	"sync"

	"golang.org/x/text/encoding/charmap"
)

const MAXMSGSIZE = 512

type MsgHandler func(Msg)

type Channel struct {
	// the mutex protects topic
	sync.Mutex
	name, topic, modes string
	names              set
	send, receive      chan Msg
	client             *Client
	started, quit      chan struct{}
	err                error
}

type Client struct {
	// the mutex protects prefix, mode, chanModes, userModes, channels, params
	sync.Mutex
	name, prefix, mode   string
	chanModes, userModes string
	nickservPass         string
	motd                 []string
	wg                   sync.WaitGroup
	socket               net.Conn
	logger               Logger
	reads, writes        chan message
	channels             map[string]*Channel
	private              *Channel
	handler              MsgHandler
	charmap              *charmap.Charmap
	params               map[string]string
	started              chan struct{}
	err                  error
	quit                 context.CancelFunc
}

type Msg interface {
	Client() *Client
	Channel() *Channel
	Text() []string
	WrappedText() []string
	Nick() string
	Prefix() string
	Messages() []message
	Logf(format string, params ...interface{})
	Debug(format string, params ...interface{})
	Reply(ctx context.Context, text []string)
	IsPrivate() bool
}

type message interface {
	Cmd() string
	Prefix() string
	Params() []string
	Nick() string
	MyNick() string
	Text() string
	Msg() Msg
	Export() []byte
	Channel() *Channel
	Client() *Client
}

// Logger should be safe to be used by several goroutines
type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
	Debug(format string, v ...interface{})
}
