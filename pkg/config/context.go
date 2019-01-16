package config

import (
	"path"

	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

const (
	contextFilename          = ".control-tower-contexts"
	ContextAlreadyExist      = cterror.Error("context already exists")
	ContextNotFound          = cterror.Error("context not found")
	NoContextInUseFoundError = cterror.Error("no context in use found")
	ContextNotSetError       = cterror.Error("context is not set")
)

type Context struct {
	Target       string `yaml:"target"`
	Team         string `yaml:"team"`
	Pipeline     string `yaml:"pipeline"`
	PipelineType string `yaml:"type"`
	InUse        bool   `yaml:"inuse"`
}

func (c *Context) Save(name string, overwrite bool) error {
	contexts, err := LoadContexts()
	if err != nil {
		return err
	}
	_, ok := contexts.Contexts[name]
	if ok {
		if !overwrite {
			return ContextAlreadyExist
		}
	}
	contexts.Contexts[name] = *c
	SaveContexts(contexts)
	return nil
}

type Contexts struct {
	Contexts map[string]Context
}

func LoadContext(name string) (Context, error) {
	contexts, err := LoadContexts()
	if err != nil {
		return Context{}, err
	}
	res, ok := contexts.Contexts[name]
	if !ok {
		return Context{}, ContextNotFound
	}
	return res, nil
}

func LoadContexts() (Contexts, error) {
	res := Contexts{Contexts: make(map[string]Context)}
	path, err := GetContextsFilepath()
	if err != nil {
		return res, err
	}
	if io.NotExist(path) {
		return res, nil
	}
	data, err := io.ReadFromFile(path)
	if err != nil {
		return res, err
	}
	err = yaml.Unmarshal(data, &res.Contexts)
	if err != nil {
		return res, err
	}
	return res, nil
}

func SaveContexts(contexts Contexts) error {
	path, err := GetContextsFilepath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(&contexts.Contexts)
	if err != nil {
		return err
	}
	err = io.WriteToFile(string(data), path)
	return err
}

func GetContextsFilepath() (string, error) {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return path.Join(home, contextFilename), nil
}

// LoadInUseContext returns all contexts, in use context name and potential error
func LoadInUseContext() (Contexts, string, error) {
	ctx, err := LoadContexts()
	if err != nil {
		return ctx, "", err
	}
	n, err := GetInUseContext(ctx)
	return ctx, n, err
}

func GetInUseContext(ctx Contexts) (string, error) {
	for n, c := range ctx.Contexts {
		if c.InUse {
			return n, nil
		}
	}
	return "", NoContextInUseFoundError
}
