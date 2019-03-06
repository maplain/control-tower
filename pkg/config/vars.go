package config

import (
	"fmt"

	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	KeyNotFoundError = cterror.Error("key not found")
)

func GetValue(path, key string) (string, error) {
	if io.NotExist(path) {
		return "", errors.Wrap(cterror.FileNotFoundError, path)
	}

	data, err := io.ReadFromFile(path)
	if err != nil {
		return "", err
	}

	d := make(map[string]interface{})
	err = yaml.Unmarshal(data, &d)
	if err != nil {
		return "", err
	}

	for k, v := range d {
		if k == key {
			return fmt.Sprintf("%v", v), nil
		}
	}
	return "", KeyNotFoundError
}

func AllKeys(path string) ([]string, error) {
	res := []string{}
	if io.NotExist(path) {
		return res, errors.Wrap(cterror.FileNotFoundError, path)
	}

	data, err := io.ReadFromFile(path)
	if err != nil {
		return res, err
	}

	d := make(map[string]interface{})
	err = yaml.Unmarshal(data, &d)
	if err != nil {
		return res, err
	}

	for k, _ := range d {
		res = append(res, k)
	}

	return res, nil
}
