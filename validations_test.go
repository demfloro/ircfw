package ircfw

import (
	"testing"
)

func TestIsASCII(t *testing.T) {
	var valid_strings = []string{
		"abcdef",
		"#testchannel",
		"&anothertestchannel",
		"ad fs",
	}
	var invalid_strings = []string{
		"игромания",
	}

	for _, valid_string := range valid_strings {
		if !isASCII(valid_string) {
			t.Fatalf("%q should be valid", valid_string)
		}
	}
	for _, invalid_string := range invalid_strings {
		if isASCII(invalid_string) {
			t.Fatalf("%q should be invalid", invalid_string)
		}
	}
}

func TestValidateChannel(t *testing.T) {
	var valid_channels = []string{
		"#mania",
		"&somes",
		"##federal",
		"#ircfw-test",
	}
	var invalid_channels = []string{
		"",
		"mania",
		"#mania,#igromania",
		"#mania #igromania",
		"#mania\x07#igromania",
		"#мания",
		"#igroмания",
		"#ccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	}

	for _, valid_channel := range valid_channels {
		if err := validateChannel(valid_channel); err != nil {
			t.Fatalf("%q should be valid, err: %q", valid_channel, err)
		}
	}
	for _, invalid_channel := range invalid_channels {
		if err := validateChannel(invalid_channel); err == nil {
			t.Fatalf("%q should be invalid", invalid_channel)
		}
	}
}
