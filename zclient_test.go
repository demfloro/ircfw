package ircfw

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"golang.org/x/text/encoding/charmap"
	"log"
	"log/syslog"
	"testing"
	"time"
)

const (
	nick     = "ircfw"
	user     = "user"
	proto    = "tcp"
	realname = "unrealname"
	server   = "irc.demsh.org:6697"
	password = ""
	jchannel = "#ircfw-test"
	timeout  = 10
)

func TestScanMsg(t *testing.T) {
	var validStream = bytes.NewReader([]byte(":irc.demsh.org 001 ircfw :Welcome to the Internet Relay Network ircfw!~ircfw@ffafb37e\r\n:irc.demsh.org 002 ircfw :Your host is irc.demsh.org, running version ngircd-26.1 (x86_64/unknown/openbsd7.0)\r\n:irc.demsh.org 003 ircfw :This server has been started Fri Oct 15 2021 at 14:01:02 (UTC)\r\n:irc.demsh.org 004 ircfw irc.demsh.org ngircd-26.1 abBcCFiIoqrRswx abehiIklmMnoOPqQrRstvVz\r\n:irc.demsh.org 005 ircfw NETWORK=ManiaNet :is my network name\r\n:ircfw!~ircfw@5838b91c MODE ircfw :+iCwx\r\n:irc.demsh.org NOTICE ircfw :Connection statistics: client 0.1 kb, server 2.2 kb.\r\n"))
	var invalidStream = bytes.NewReader([]byte(":irc.demsh.org 001 \rircfw :Welcome to the Internet Relay Network ircfw!~ircfw@ffafb37e:irc.demsh.org 002 ircfw :Your host is irc.demsh.org, running version ngircd-26.1 (x86_64/unknown/openbsd7.0):irc.demsh.org 003 ircfw :This server has been started Fri Oct 15 2021 at 14:01:02 (UTC):irc.demsh.org 004 ircfw irc.demsh.org ngircd-26.1 abBcCFiIoqrRswx abehiIklmMnoOPqQrRstvVz:irc.demsh.org 005 ircfw NETWORK=ManiaNet :is my network name:ircfw!~ircfw@5838b91c MODE ircfw :+iCwx:irc.demsh.org NOTICE ircfw :Connection statistics: client 0.1 kb, server 2.2 kb.\n\r\n"))

	scanner := bufio.NewScanner(validStream)
	scanner.Buffer(make([]byte, MAXMSGSIZE), MAXMSGSIZE)
	scanner.Split(scanMsg)

	invalidScanner := bufio.NewScanner(invalidStream)
	invalidScanner.Buffer(make([]byte, MAXMSGSIZE), MAXMSGSIZE)
	invalidScanner.Split(scanMsg)

	for scanner.Scan() {
		scanner.Text()
	}
	err := scanner.Err()
	if err != nil {
		t.Fatalf("%#v", err)
	}

	for invalidScanner.Scan() {
		invalidScanner.Text()
	}
	err = invalidScanner.Err()
	if err != nil {
		if !errors.Is(err, bufio.ErrTooLong) {
			t.Fatalf("%#v", err)
		}
	}
}

type testLogger struct {
	debug  bool
	logger *log.Logger
}

func (t testLogger) Log(v ...interface{}) {
	t.logger.Print(v...)
}

func (t testLogger) Logf(format string, v ...interface{}) {
	t.logger.Printf(format, v...)
}

func (t testLogger) Debug(format string, v ...interface{}) {
	if !t.debug {
		return
	}
	t.logger.Printf(format, v...)
}

func newLogger(debug bool, logger *log.Logger) (out testLogger, err error) {
	var level syslog.Priority
	if debug {
		level = syslog.LOG_DEBUG
	} else {
		level = syslog.LOG_ERR
	}
	if logger == nil {
		logger, err = syslog.NewLogger(level, 0)
		if err != nil {
			return
		}
	}
	out.debug = debug
	out.logger = logger
	return
}

func TestNewClient(t *testing.T) {
	socket, err := tls.Dial(proto, server, nil)
	if err != nil {
		t.Fatal(err)
	}
	logger, err := newLogger(true, nil)
	if err != nil {
		t.Fatal(err)
	}
	//mock := newMockProxy(socket, logger)
	//client := NewClient(nick, user, realname, password, mock, logger)
	charmap := charmap.Windows1251
	rootCtx := context.Background()
	client, cancelClient := NewClient(rootCtx, nick, user, realname, password, "", socket, logger, drainHandler, charmap)
	defer cancelClient()
	ctx, cancel := context.WithTimeout(rootCtx, 5*time.Second)
	_, err = client.Join(ctx, jchannel)
	cancel()
	if err != nil {
		logger.Log(err)
	}
	client.Wait()
}

func drainHandler(m Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := m.Client()
	channel := m.Channel()
	for _, line := range m.Text() {
		if len(line) > 5 && lowcase(line[:5]) == "!say " {
			m.Reply(ctx, []string{line[5:]})
			return
		}
		if lowcase(line) == "!part" {
			channel.Part()
			return
		}
		if lowcase(line) == "!quit" {
			client.Quit("Requested by privmsg")
		}
		if lowcase(line) == "!status" {
			return
		}
		if lowcase(line) == "!topic" {
			channel.queryTopic()
		}
		if len(line) > 7 && lowcase(line[:7]) == "!whois " {
			client.Whois(line[7:])
		}
		if len(line) > 6 && lowcase(line[:6]) == "!join " {
			client.Join(ctx, line[6:])
		}
	}
}
