package io

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/olekukonko/tablewriter"
	yaml "gopkg.in/yaml.v2"
)

const (
	UnsupportedOutputFormatError = cterror.Error("unsupported output type")
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

type printer struct {
	format string
}

type Printer interface {
	Display(bool, [][]string, ...[]string)
}

func NewPrinter(outputFormat string) (*printer, error) {
	switch outputFormat {
	case "table", "csv":
	default:
		return nil, errors.Wrap(UnsupportedOutputFormatError, outputFormat)
	}
	return &printer{outputFormat}, nil
}

func (p *printer) Display(header bool, data [][]string, headers ...[]string) {
	switch p.format {
	case "table":
		writeTable(header, data, headers...)
	case "csv":
		if header {
			for _, h := range headers {
				fmt.Println(strings.Join(h, ","))
			}
		}
		for _, d := range data {
			fmt.Println(strings.Join(d, ","))
		}
	}
}

func writeTable(header bool, data [][]string, headers ...[]string) {
	table := tablewriter.NewWriter(os.Stdout)
	if header && len(headers) > 0 {
		table.SetHeader(headers[0])
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
