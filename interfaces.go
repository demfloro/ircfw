package ircfw

import (
	"context"
	"net"
	"sync"
	"time"

	"golang.org/x/text/encoding/charmap"
	"gopkg.in/tomb.v2"
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
	tomb          *tomb.Tomb
	socket        net.Conn
	reads, writes chan message
	private       *Channel
	started       chan struct{}
	handler       MsgHandler
	charmap       *charmap.Charmap
	logger        Logger
	aliveTimeout  time.Duration
	sync.Mutex
	// fields below are protected by the mutex
	lastMessage          time.Time
	name, prefix, mode   string
	chanModes, userModes string
	nickservPass         string
	motd                 []string
	channels             map[string]*Channel
	params               map[string]string
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
	Deadline() time.Time
}

// Logger should be safe to be used by several goroutines
type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
	Debug(format string, v ...interface{})
}
