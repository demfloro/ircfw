package ircfw

import (
	"golang.org/x/text/encoding/charmap"
)

func encode(input string, charmap *charmap.Charmap) (result []byte) {
	var b byte

	for _, r := range input {
		b, _ = charmap.EncodeRune(r)
		result = append(result, b)
	}
	return
}

func decode(input []byte, charmap *charmap.Charmap) (result string) {
	var (
		runes []rune
		Rune  rune
	)

	for _, v := range input {
		Rune = charmap.DecodeByte(v)
		runes = append(runes, Rune)
	}
	result = string(runes)
	return
}
