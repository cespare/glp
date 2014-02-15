package main

import (
	"sort"
)

// An smap is a sorted map of strings.
type smap []string

func newSmap() smap { return nil }

func (m *smap) Add(s string) {
	i, ok := m.find(s)
	if !ok {
		*m = append((*m)[:i], append(smap{s}, (*m)[i:]...)...)
	}
}

func (m *smap) Contains(s string) bool {
	_, ok := m.find(s)
	return ok
}

func (m *smap) Remove(s string) {
	i, ok := m.find(s)
	if ok {
		*m = append((*m)[:i], (*m)[i+1:]...)
	}
}

func (m *smap) find(s string) (i int, ok bool) {
	i = sort.SearchStrings([]string(*m), s)
	ok = i < len(*m) && (*m)[i] == s
	return i, ok
}
