package io

type StringSet map[string]struct{}

func NewStringSet() StringSet {
	return StringSet(make(map[string]struct{}))
}

func (s StringSet) Add(key string) {
	s[key] = struct{}{}
}

func (s StringSet) Remove(key string) {
	delete(s, key)
}
