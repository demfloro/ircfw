package ircfw

import (
	"fmt"
	"strings"
	"sync"
)

type set struct {
	sync.Mutex
	s map[string]struct{}
}

func NewSet() set {
	return set{
		s: make(map[string]struct{}),
	}
}

func (s *set) Has(k string) bool {
	s.Lock()
	_, ok := s.s[k]
	s.Unlock()
	return ok
}

func (s *set) Add(k string) {
	s.Lock()
	s.s[k] = struct{}{}
	s.Unlock()
}

func (s *set) Remove(k string) {
	s.Lock()
	delete(s.s, k)
	s.Unlock()
}

func (s *set) Replace(olde string, newe string) {
	s.Lock()
	_, ok := s.s[olde]
	if ok {
		delete(s.s, olde)
	}
	s.s[newe] = struct{}{}
	s.Unlock()
}

func (s *set) Clear() {
	s.Lock()
	s.s = make(map[string]struct{})
	s.Unlock()
}

func (s *set) Size() int {
	s.Lock()
	defer s.Unlock()
	return len(s.s)
}

func (s *set) String() string {
	var slice []string
	s.Lock()
	defer s.Unlock()
	for k, _ := range s.s {
		slice = append(slice, k)
	}
	return fmt.Sprintf("set(%q)", strings.Join(slice, ", "))
}
