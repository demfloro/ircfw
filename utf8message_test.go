package ircfw

import (
	"testing"
)

func TestParseParams(t *testing.T) {
	samples := []string{
		"NETWORK=ManiaNet :is my network name",
		"RFC2812 IRCD=ngIRCd CHARSET=UTF-8 CASEMAPPING=ascii PREFIX=(qaohv)~&@%+ CHANTYPES=#&+ CHANMODES=beI,k,l,imMnOPQRstVz CHANLIMIT=#&+:10 :are supported on this server",
		"CHANNELLEN=50 NICKLEN=9 TOPICLEN=490 AWAYLEN=127 KICKLEN=400 MODES=5 MAXLIST=beI:50 EXCEPTS=e INVEX=I PENALTY FNC :are supported on this server",
		"hello",
	}
	valid_results := [][]string{
		[]string{
			"NETWORK=ManiaNet",
			"is my network name",
		},
		[]string{"RFC2812",
			"IRCD=ngIRCd",
			"CHARSET=UTF-8",
			"CASEMAPPING=ascii",
			"PREFIX=(qaohv)~&@%+",
			"CHANTYPES=#&+",
			"CHANMODES=beI,k,l,imMnOPQRstVz",
			"CHANLIMIT=#&+:10",
			"are supported on this server",
		},
		[]string{
			"CHANNELLEN=50",
			"NICKLEN=9",
			"TOPICLEN=490",
			"AWAYLEN=127",
			"KICKLEN=400",
			"MODES=5",
			"MAXLIST=beI:50",
			"EXCEPTS=e",
			"INVEX=I",
			"PENALTY",
			"FNC",
			"are supported on this server",
		},
		[]string{
			"hello",
		},
	}

	for i, sample := range samples {
		result := parseParams(sample)
		for j, s := range result {
			if s != valid_results[i][j] {
				t.Fatalf("%#v != %#v", result, valid_results[i])
			}
		}
	}
}

func TestParseMsg(t *testing.T) {
	samples := []string{
		":demsh!~demsh@12a8e790 PRIVMSG #ircfw-test :heyo people!",
		":demsh!~demsh@e20eb9ad PRIVMSG #ircfw-test :don't you dare to visit https://demsh.org/",
		":ircfw!~ircfw@5838b91c MODE ircfw :+iCwx",
		":irc.demsh.org 001 ircfw :Welcome to the Internet Relay Network ircfw!~ircfw@ffafb37e",
		":irc.demsh.org 004 ircfw irc.demsh.org ngircd-26.1 abBcCFiIoqrRswx abehiIklmMnoOPqQrRstvVz",
		"PING :irc.demsh.org",
	}
	valids := []utf8message{
		utf8message{
			prefix: "demsh!~demsh@12a8e790",
			cmd:    "PRIVMSG",
			params: []string{
				"#ircfw-test",
				"heyo people!",
			},
		},
		utf8message{
			prefix: "demsh!~demsh@e20eb9ad",
			cmd:    "PRIVMSG",
			params: []string{
				"#ircfw-test",
				"don't you dare to visit https://demsh.org/",
			},
		},
		utf8message{
			prefix: "ircfw!~ircfw@5838b91c",
			cmd:    "MODE",
			params: []string{
				"ircfw",
				"+iCwx",
			},
		},
		utf8message{
			prefix: "irc.demsh.org",
			cmd:    "001",
			params: []string{
				"ircfw",
				"Welcome to the Internet Relay Network ircfw!~ircfw@ffafb37e",
			},
		},
		utf8message{
			prefix: "irc.demsh.org",
			cmd:    "004",
			params: []string{
				"ircfw",
				"irc.demsh.org",
				"ngircd-26.1",
				"abBcCFiIoqrRswx",
				"abehiIklmMnoOPqQrRstvVz",
			},
		},
		utf8message{
			prefix: "",
			cmd:    "PING",
			params: []string{
				"irc.demsh.org",
			},
		},
	}
	for i, sample := range samples {
		valid := valids[i]
		msg, err := parseUTF8Message([]byte(sample), nil)
		if err != nil {
			t.Fatalf("Failed to parse: %#v, err: %#v", sample, err)
		}
		if msg.Prefix() != valid.Prefix() || msg.Cmd() != valid.Cmd() {
			t.Fatalf("Invalid parse: %#v != %#v", msg, valid)
		}
		for j, param := range msg.Params() {
			if param != valid.Params()[j] {
				t.Fatalf("Invalid parse: %#v != %#v", msg, valid)
			}
		}
	}
}

func TestExportMsg(t *testing.T) {
	samples := []utf8message{
		utf8message{
			prefix: "demsh!~demsh@12a8e790",
			cmd:    "PRIVMSG",
			params: []string{"#ircfw-test", "heyo people!"},
			client: nil,
		},
		utf8message{
			prefix: "",
			cmd:    "PING",
			params: []string{"irc.demsh.org"},
			client: nil,
		},
		utf8message{
			prefix: "",
			cmd:    "PONG",
			params: []string{"irc.demsh.org"},
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
