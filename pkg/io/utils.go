package io

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

type Remover func()

// WriteToTempFile will create a temporary file, write data in it and return
// tmp file's name, a lambda which removes this temporary file and an error
// caller needs to call Remover() after reading tmpfile
func WriteToTempFile(data string, dir, prefix string) (string, Remover, error) {
	tmpfile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return "", nil, err
	}
	// clean up
	remover := func() {
		os.Remove(tmpfile.Name())
	}

	err = WriteToFile(data, tmpfile.Name())
	if err != nil {
		return "", nil, err
	}

	err = tmpfile.Close()
	if err != nil {
		return "", nil, err
	}
	return tmpfile.Name(), remover, nil
}
