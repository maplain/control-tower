package io

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	yaml "gopkg.in/yaml.v2"
)

func GetFilenames(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path != root {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func WriteTable(data [][]string, header ...[]string) {
	table := tablewriter.NewWriter(os.Stdout)
	if len(header) > 0 {
		table.SetHeader(header[0])
	}
	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
}

func ReadFromFile(file string) ([]byte, error) {
	data, err := ioutil.ReadFile(file)
	return data, err
}

func DumpYaml(data string) (string, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return "", err
	}
	d, err := yaml.Marshal(&m)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func NotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func BinaryPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	return path, nil
}

func WriteToFile(data, file string) error {
	err := os.MkdirAll(path.Dir(file), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, []byte(data), os.ModePerm)
}

type Values map[string]string

func (v Values) Get(key string) (string, bool) {
	res, ok := v[key]
	return res, ok
}

func (v Values) GetInt(key string) (int, error) {
	value, ok := v[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("key %s does not exist", key))
	}
	i, err := strconv.Atoi(value)
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
	for _, name := range ordered {
		value := inputs[name]
	Setvalue:
		if value != "" {
			fmt.Printf("type in the value for %s (type Enter to use default: %s)\n", name, value)
		} else {
			fmt.Printf("type in the value for %s\n", name)
		}
		v, _ := reader.ReadString('\n')
		v = strings.TrimSpace(v)
		if v != "" {
			inputs[name] = v
		}
		if v == "" && value == "" {
			goto Setvalue
		}
	}
	return inputs
}
