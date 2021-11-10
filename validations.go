package ircfw

import (
	"errors"
	"strings"
	"unicode"
)

const (
	CHAN_LENGTH_LIMIT = 200
)

// https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func isASCII(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func hasNULL(s string) bool {
	return strings.Contains(s, "\x00")
}

// https://datatracker.ietf.org/doc/html/rfc1459#section-1.3
func validateChannel(channel string) error {
	if len(channel) == 0 {
		return errors.New("empty")
	}
	if len(channel) > CHAN_LENGTH_LIMIT {
		return errors.New("longer than 200 bytes")
	}
	if !isASCII(channel) {
		return errors.New("non-ASCII")
	}
	if strings.ContainsAny(channel, ", \x00\x07") {
		return errors.New("illegal symbol")
	}
	if !(strings.HasPrefix(channel, "#") || strings.HasPrefix(channel, "&")) {
		return errors.New("does not start with # or &")
	}
	return nil
}

func isChannel(channel string) bool {
	if err := validateChannel(channel); err != nil {
		return false
	}
	return true
}

func validate(s string) error {
	if strings.ContainsAny(s, "\x00\x07") {
		return errors.New("illegal symbol")
	}
	return nil
}

func validateNick(nick string) error {
	if len(nick) == 0 {
		return errors.New("empty")
	}
	if len(nick) > 9 {
		return errors.New("longer than 9 bytes")
	}
	if !isASCII(nick) {
		return errors.New("non-ASCII")
	}
	if strings.ContainsAny(nick, ", \x00\x07") {
		return errors.New("illegal symbol")
	}
	if strings.HasPrefix(nick, "#") || strings.HasPrefix(nick, "&") {
		return errors.New("starts with # or &")
	}
	return nil
}

func isNick(nick string) bool {
	if err := validateNick(nick); err != nil {
		return false
	}
	return true
}
