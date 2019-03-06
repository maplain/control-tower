package io

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type UserInput struct {
	Description string
	Value       string
}
type Values map[string]UserInput

func NewValues() Values {
	return Values(make(map[string]UserInput))
}

func NewValuesFromStringSlice(s []string) Values {
	res := make(map[string]UserInput)
	for _, str := range s {
		res[str] = UserInput{}
	}
	return Values(res)
}

func FromMap(m map[string]string) Values {
	res := make(map[string]UserInput)
	for k, v := range m {
		res[k] = UserInput{Value: v}
	}
	return Values(res)
}

func (v Values) ToMap() map[string]string {
	res := make(map[string]string)
	for k, val := range v {
		res[k] = val.Value
	}
	return res
}

func (v Values) AddValues(values Values) {
	for k, val := range values {
		v[k] = val
	}
}
func (v Values) Add(key, val string) {
	ori, found := v[key]
	if !found {
		v[key] = UserInput{Value: val}
	} else {
		ori.Value = val
		v[key] = ori
	}
}

func (v Values) GetValue(key string) (string, bool) {
	res, ok := v[key]
	return res.Value, ok
}

func (v Values) GetInt(key string) (int, error) {
	value, ok := v[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("key %s does not exist", key))
	}
	i, err := strconv.Atoi(value.Value)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("cannot convert %s to integer", value))
	}
	return i, nil
}

func InteractivePopulateStringValues(inputs Values) Values {
	reader := bufio.NewReader(os.Stdin)
	var ordered []string
	for k, _ := range inputs {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)

	res := NewValues()

	for _, name := range ordered {
		value := inputs[name]
		if strings.TrimSpace(value.Description) != "" {
			fmt.Println(value.Description)
		}
	Setvalue:
		if value.Value != "" {
			fmt.Printf("type in the value for %s (type Enter to use default: %s)\n", name, value.Value)
		} else {
			fmt.Printf("type in the value for %s\n", name)
		}
		v, _ := reader.ReadString('\n')
		v = strings.TrimSpace(v)
		if v != "" {
			res.Add(name, v)
		}
		if v == "" && value.Value == "" {
			goto Setvalue
		}
	}
	return res
}
