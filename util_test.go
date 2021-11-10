package ircfw

import (
	"strings"
	"testing"
)

const (
	MLIMIT = 64
)

func TestSplitByLen(t *testing.T) {
	var lines = []string{
		"",
		"abcd",
		"abcdef",
		"abdsd ssddc dsdsdadaadwdwwdef",
		"aaaaabbbdd dfsdfsdf adadsd",
		"addddddddddddddddddddd asssss ss ssssssssssssssssaaa aaa ddddddddddddddddd sssssssssssssss",
		"The quick brown fox jumps over the lazy dog",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		"Go - компилируемый многопоточный язык программирования, разработанный внутри компании Google. Разработка Go началась в сентябре 2007 года, его непосредственным проектированием занимались Роберт Гризмер, Роб Пайк и Кен Томпсон, занимавшиеся до этого проектом разработки операционной системы Inferno. Официально язык был представлен в ноябре 2009 года. На данный момент поддержка официального компилятора, разрабатываемого создателями языка, осуществляется для операционных систем FreeBSD, OpenBSD, Linux, macOS, Windows, DragonFly BSD, Plan 9, Solaris, Android, AIX. Также Go поддерживается набором компиляторов gcc, существует несколько независимых реализаций. Ведётся разработка второй версии языка",
	}

	for _, line := range lines {
		result := splitByLen(line, MLIMIT, 0)
		for _, subline := range result {
			if len(subline) > MLIMIT {
				t.Fatalf("%q != %q", line, result)
			}
		}
		if strings.Join(result, " ") != line {
			t.Fatalf("%q != %q", line, result)
		}
	}
}
