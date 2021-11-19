package ircfw

import (
	"testing"
	"time"

	"golang.org/x/text/encoding/charmap"
)

func TestParseByteMessage(t *testing.T) {
	samples := []string{
		":demsh!~demsh@12a8e790 PRIVMSG #ircfw-test :heyo people!",
		":demsh!~demsh@e20eb9ad PRIVMSG #ircfw-test :don't you dare to visit https://demsh.org/",
		":ircfw!~ircfw@5838b91c MODE ircfw :+iCwx",
		":irc.demsh.org 001 ircfw :Welcome to the Internet Relay Network ircfw!~ircfw@ffafb37e",
		":irc.demsh.org 004 ircfw irc.demsh.org ngircd-26.1 abBcCFiIoqrRswx abehiIklmMnoOPqQrRstvVz",
		"PING :irc.demsh.org",
	}
	client := Client{charmap: charmap.Windows1251}
	valids := []bytemessage{
		bytemessage{
			prefix: []byte("demsh!~demsh@12a8e790"),
			cmd:    []byte("PRIVMSG"),
			params: [][]byte{
				[]byte("#ircfw-test"),
				[]byte("heyo people!"),
			},
			client: &client,
		},
		bytemessage{
			prefix: []byte("demsh!~demsh@e20eb9ad"),
			cmd:    []byte("PRIVMSG"),
			params: [][]byte{
				[]byte("#ircfw-test"),
				[]byte("don't you dare to visit https://demsh.org/"),
			},
			client: &client,
		},
		bytemessage{
			prefix: []byte("ircfw!~ircfw@5838b91c"),
			cmd:    []byte("MODE"),
			params: [][]byte{
				[]byte("ircfw"),
				[]byte("+iCwx"),
			},
			client: &client,
		},
		bytemessage{
			prefix: []byte("irc.demsh.org"),
			cmd:    []byte("001"),
			params: [][]byte{
				[]byte("ircfw"),
				[]byte("Welcome to the Internet Relay Network ircfw!~ircfw@ffafb37e"),
			},
			client: &client,
		},
		bytemessage{
			prefix: []byte("irc.demsh.org"),
			cmd:    []byte("004"),
			params: [][]byte{
				[]byte("ircfw"),
				[]byte("irc.demsh.org"),
				[]byte("ngircd-26.1"),
				[]byte("abBcCFiIoqrRswx"),
				[]byte("abehiIklmMnoOPqQrRstvVz"),
			},
			client: &client,
		},
		bytemessage{
			prefix: []byte{},
			cmd:    []byte("PING"),
			params: [][]byte{
				[]byte("irc.demsh.org"),
			},
			client: &client,
		},
	}
	for i, sample := range samples {
		valid := valids[i]
		msg, err := parseByteMessage([]byte(sample), time.Time{}, &client)
		if err != nil {
			t.Fatalf("Failed to parse: %#v, err: %q", sample, err)
		}
		if msg.Prefix() != valid.Prefix() || msg.Cmd() != valid.Cmd() {
			t.Logf("%q", sample)
			t.Logf("%q %q", valid.Prefix(), valid.Cmd())
			t.Logf("%q %q", msg.Prefix(), msg.Cmd())
			t.Fatalf("Invalid parse: %#v != %#v", msg, valid)
		}
		for j, param := range msg.Params() {
			if param != valid.Params()[j] {
				t.Fatalf("Invalid parse: %#v != %#v", msg, valid)
			}
		}
	}
}

func TestExportByteMessage(t *testing.T) {
	samples := []bytemessage{
		bytemessage{
			prefix: []byte("demsh!~demsh@12a8e790"),
			cmd:    []byte("PRIVMSG"),
			params: [][]byte{[]byte("#ircfw-test"), []byte("heyo people!")},
			client: nil,
		},
		bytemessage{
			prefix: []byte{},
			cmd:    []byte("PING"),
			params: [][]byte{[]byte("irc.demsh.org")},
			client: nil,
		},
		bytemessage{
			prefix: []byte{},
			cmd:    []byte("PONG"),
			params: [][]byte{[]byte("irc.demsh.org")},
			client: nil,
		},
	}
	corrects := []string{
		"PRIVMSG #ircfw-test :heyo people!\r\n",
		"PING :irc.demsh.org\r\n",
		"PONG :irc.demsh.org\r\n",
	}
	for i, sample := range samples {
		if export := sample.Export(); string(export) != corrects[i] {
			t.Fatalf("%#v != %#v", string(export), corrects[i])
		}
	}
}
