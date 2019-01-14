package config

import (
	"sort"
	"strings"

	"github.com/maplain/control-tower/pkg/io"
)

type Tags struct {
	io.StringSet
}

func NewTags() Tags {
	return Tags{io.NewStringSet()}
}

func (t Tags) String() string {
	res := []string{}
	for k, _ := range t.StringSet {
		res = append(res, k)
	}
	sort.Strings(res)
	return strings.Join(res, ",")
}
