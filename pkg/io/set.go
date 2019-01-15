package io

type StringSet map[string]struct{}

func NewStringSet() StringSet {
	return StringSet(make(map[string]struct{}))
}

func NewStringSetFromSlice(s []string) StringSet {
	res := NewStringSet()
	for _, str := range s {
		res.Add(str)
	}
	return res
}

func (s StringSet) Add(key string) {
	s[key] = struct{}{}
}

func (s StringSet) Remove(key string) {
	delete(s, key)
}
