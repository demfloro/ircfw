package ircfw

import (
	"context"
	"net"

	"golang.org/x/text/encoding/charmap"
)

type Option func(*config)

type config struct {
	nick, ident, realName  string
	password, nickservPass string
	handler                MsgHandler
	socket                 net.Conn
	logger                 Logger
	charmap                *charmap.Charmap
	context                context.Context
}

func defaultConfig() config {
	return config{
		nick:     "ircfw",
		ident:    "ircfw",
		realName: "ircfw",
		context:  context.Background(),
	}
}

func Context(ctx context.Context) Option {
	return func(c *config) {
		c.context = ctx
	}
}

func Charmap(charmap *charmap.Charmap) Option {
	return func(c *config) {
		c.charmap = charmap
	}
}

func SetLogger(logger Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

func Socket(socket net.Conn) Option {
	return func(c *config) {
		c.socket = socket
	}
}

func Handler(handler MsgHandler) Option {
	return func(c *config) {
		c.handler = handler
	}
}

func Nick(nick string) Option {
	return func(c *config) {
		c.nick = nick
	}
}

func Ident(ident string) Option {
	return func(c *config) {
		c.ident = ident
	}
}

func RealName(realname string) Option {
	return func(c *config) {
		c.realName = realname
	}
}

func NickServPass(password string) Option {
	return func(c *config) {
		c.nickservPass = password
	}
}

func Password(password string) Option {
	return func(c *config) {
		c.password = password
	}
}
