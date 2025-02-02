package ircfw

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/tomb.v2"
)

type handler func(message)

var (
	handlers = map[string]handler{
		"PING":    handlePing,
		"PONG":    handlePong,
		"PRIVMSG": handlePrivmsg,
		"NOTICE":  handleNotice,
		"ERROR":   handleError,
		"JOIN":    handleJoin,
		"NICK":    handleNick,
		"PART":    handlePart,
		"MODE":    handleMode,
		"001":     handleWelcome,
		"004":     handleMyInfo,
		"005":     handleISupport,
		"275":     handleWhois,
		"311":     handleWhois,
		"312":     handleWhois,
		"319":     handleWhois,
		"332":     handleTopic,
		"333":     handleTopic,
		"353":     handleNames,
		"372":     handleMOTD,
		"396":     handleHostname,
		"473":     handleJoinError,
	}
)

// Meant to run in separate goroutine
func (c *Client) serveLoop() error {
	for {
		select {
		case <-c.tomb.Dying():
			c.Debug("serveLoop dying")
			return tomb.ErrDying
		case msg, open := <-c.reads:
			if !open {
				c.Debug("c.reads closed, quitting")
				return ErrReadsClosed
			}
			handler, exists := handlers[msg.Cmd()]
			if exists {
				handler(msg)
			} else {
				logHandler(msg)
			}
		}
	}
}

func handleJoinError(msg message) {
	client := msg.Client()
	if msg.Cmd() != "473" {
		client.Debug("Got error: %#v", msg)
		return
	}
	params := msg.Params()
	chanName := params[1]
	err := fmt.Errorf("%q: %q", chanName, params[len(params)-1])
	if channel := client.fetchChannel(chanName); channel != nil {
		client.Lock()
		delete(client.channels, chanName)
		client.Unlock()
		channel.err = err
		channel.kill()
	}
}

func handleMyInfo(msg message) {
	params := msg.Params()
	client := msg.Client()
	if len(params) == 5 {
		client.Lock()
		client.userModes = params[3]
		client.chanModes = params[4]
		client.Unlock()
		return
	}
	client.Debug("RPL_MYINFO param length is not 5")
}

func handleWhois(msg message) {
}

func handlePong(msg message) {
}

func logHandler(msg message) {
	msg.Client().Debug("Unhandled: %#v", msg)
}

func handleMOTD(msg message) {
	motd := strings.TrimSpace(strings.Join(msg.Params()[1:], " "))
	client := msg.Client()
	client.Lock()
	client.motd = append(msg.Client().motd, motd)
	client.Unlock()
}

func handleWelcome(msg message) {
	client := msg.Client()
	params := msg.Params()
	paramSlice := strings.Split(params[len(params)-1], " ")
	client.Lock()
	client.prefix = paramSlice[len(paramSlice)-1]
	client.Unlock()
}

func handleModeNick(msg message) {
	client := msg.Client()
	params := msg.Params()
	target, mode := params[0], params[1]
	client.UpdateMode(target, mode)
	if client.nickservPass == "" {
		return
	}
	client.sendMessage("PRIVMSG", []string{"NickServ", "identify " + client.nickservPass})
}

func handleModeChannel(msg message) {
	channel := msg.Channel()
	channel.modes = msg.Params()[1]
}

func handleMode(msg message) {
	params := msg.Params()
	if len(params) < 2 {
		msg.Client().Logf("Got MODE with less than 2 parameters: %#v", msg)
		return
	}
	if isNick(params[0]) {
		handleModeNick(msg)
		return
	}
	if isChannel(params[0]) {
		handleModeChannel(msg)
	}
}

func handleHostname(msg message) {
	hostname := msg.Params()[1]
	msg.Client().setHostname(hostname)
}

func handleNotice(msg message) {
	msg.Client().Debug("Notice from %q: %q", msg.Nick(), msg.Text())
}

func handlePing(msg message) {
	msg.Client().pong(msg.Params())
}

func send(ctx context.Context, msg Msg, channel chan<- Msg) {
	select {
	case channel <- msg:
	case <-ctx.Done():
		msg.Debug("error sending message: %q", ctx.Err())
	}
}

func handlePrivmsgPrivate(msg message) {
	client := msg.Client()
	ctx := client.tomb.Context(nil)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	send(ctx, msg.Msg(), client.private.receive)
	cancel()
}

func handlePrivmsg(msg message) {
	chanName := msg.Params()[0]
	client := msg.Client()
	if isNick(chanName) {
		handlePrivmsgPrivate(msg)
		return
	}
	if !isChannel(chanName) {
		client.Debug("Got PRIVMSG for invalid channel %q: %#v", chanName, msg)
		return
	}
	channel := msg.Channel()
	if channel == nil {
		client.Debug("Got unexpected PRIVMSG for %q: %#v", chanName, msg)
		return
	}
	ctx := client.tomb.Context(nil)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	send(ctx, msg.Msg(), channel.receive)
	cancel()
}

func handleNick(msg message) {
	oldnick := msg.Nick()
	newnick := msg.Params()[0]
	if msg.Client().Nick() == oldnick {
		msg.Client().setNick(newnick)
		return
	}
	msg.Client().Lock()
	defer msg.Client().Unlock()
	for _, channel := range msg.Client().channels {
		channel.names.Replace(oldnick, newnick)
	}
}

func handleError(msg message) {
	msg.Client().Logf("Error from server: %#v", msg)
}

func handleISupport(msg message) {
	client := msg.Client()
	// Allow to JOIN channels
	safeClose(client.started)
	client.Lock()
	defer client.Unlock()
	for _, param := range msg.Params() {
		splitted := strings.Split(param, "=")
		if len(splitted) == 2 {
			msg.Client().params[splitted[0]] = splitted[1]
			continue
		}
		if !strings.Contains(splitted[0], " ") {
			msg.Client().params[splitted[0]] = ""
		}
	}
}

func handleJoin(msg message) {
	msgnick := msg.Nick()
	mynick := msg.MyNick()
	chanName := msg.Params()[0]
	if mynick == msgnick {
		// Can't use message.Channel() here
		msg.Client().Lock()
		if channel := msg.Client().fetchChannel(chanName); channel != nil {
			channel.start()
		} else {
			msg.Client().Debug("Unsolicited JOIN for %q", chanName)
		}
		msg.Client().Unlock()
		return
	}
	if channel := msg.Channel(); channel != nil {
		channel.names.Add(msgnick)
		return
	}
	msg.Client().Debug("Got unsolicited notification about join: %#v", msg)
}

func handleTopic(msg message) {
	topic := msg.Params()[2]
	channel := msg.Channel()
	if channel == nil {
		return
	}
	channel.Lock()
	channel.setTopic(topic)
	channel.Unlock()
}

func handleNames(msg message) {
	channel := msg.Channel()
	for _, nick := range strings.Split(msg.Params()[3], " ") {
		channel.names.Add(nick)
	}
}

func handlePart(msg message) {
	channel := msg.Channel()
	if msg.Nick() == msg.MyNick() {
		client := msg.Client()
		client.Lock()
		delete(client.channels, channel.name)
		channel.kill()
		client.Unlock()
		return
	}
	channel.names.Remove(msg.Nick())
}
