package ircfw

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

const (
	RecursionLimit = 1000
)

func isUTF8(data []byte) bool {
	return !strings.ContainsRune(string(data), utf8.RuneError)
}

func safeClose(c chan struct{}) {
	select {
	case <-c:
		return
	default:
		close(c)
	}
}

func splitByLen(line string, limit int, depth uint) (result []string) {
	var color = NoColor
	depth++
	if depth == RecursionLimit {
		return
	}
	if limit == 0 || len(line) == 0 {
		return
	}
	line = strings.TrimSpace(line)
	if string(line[0]) == ColorTag {
		color = lookupColor(string(line[1:3]))
		if limit < 3 {
			return
		}
		limit -= 2
	}
	length := len(line)
	if length > limit {
		i := strings.LastIndex(line[:limit], " ")
		subline := ""
		if i == -1 {
			i = limit
		}
		result = append(result, line[:i])
		if color == NoColor {
			subline = strings.TrimSpace(line[i:])
		} else {
			subline = ColorTag + color.String() + strings.TrimSpace(line[i:])
		}
		if color != NoColor {
			limit += 2
		}
		if len(subline) > limit {
			result = append(result, splitByLen(subline, limit, depth)...)
		} else {
			result = append(result, subline)
		}
	} else {
		result = append(result, line)
	}
	return
}

func pop(line string, separator string) (string, string) {
	splitted := strings.SplitN(line, separator, 2)
	if len(splitted) == 2 {
		return splitted[0], splitted[1]
	}
	return splitted[0], ""
}

func join(lines []string, separator string) string {
	return strings.Join(lines, separator)
}

func textLen(lines []string) (result int) {
	for _, line := range lines {
		result += len(line)
	}
	return
}

func bytepop(line []byte, separator []byte) ([]byte, []byte) {
	splitted := bytes.SplitN(line, separator, 2)
	if len(splitted) == 2 {
		return splitted[0], splitted[1]
	}
	return splitted[0], []byte{}
}

func dropCRLF(data []byte) []byte {
	if len(data) > 1 && data[len(data)-2] == '\r' && data[len(data)-1] == '\n' {
		return data[0 : len(data)-2]
	}
	return data
}

//Split function for Scanner interface
func scanMsg(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte("\r\n")); i > 0 {
		return i + 2, dropCRLF(data[0:i]), nil
	}
	if atEOF {
		return len(data), dropCRLF(data), nil
	}
	return 0, nil, nil
}

func lowcase(s string) string {
	return strings.ToLower(s)
}
