package config

import (
	"sort"
	"strings"
)

type Tag string
type Tags map[Tag]struct{}

func NewTags() Tags {
	return make(map[Tag]struct{})
}

func (t Tags) Add(tag string) {
	t[Tag(tag)] = struct{}{}
}

func (t Tags) Remove(tag string) {
	delete(t, Tag(tag))
}

func (t Tags) String() string {
	res := []string{}
	for k, _ := range t {
		res = append(res, string(k))
	}
	sort.Strings(res)
	return strings.Join(res, ",")
}
