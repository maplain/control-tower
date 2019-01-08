package templates

import (
	"io/ioutil"
	"regexp"
	"strings"

	boshtemplate "github.com/cloudfoundry/bosh-cli/director/template"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	TemplateFileNotFoundError              = cterror.Error("template file not found")
	CouldNotReadTemplateVariablesFileError = cterror.Error("could not read template variables file")
)

func Interpolate(configFile string, varFiles []string) (string, error) {
	config, err := io.ReadFromFile(configFile)
	if err != nil {
		return "", err
	}

	var params []boshtemplate.Variables
	for _, path := range varFiles {
		templateVars, err := ioutil.ReadFile(path)
		if err != nil {
			return "", errors.Wrap(errors.Wrap(err, path), "during interpolation")
		}

		var staticVars boshtemplate.StaticVariables
		err = yaml.Unmarshal(templateVars, &staticVars)
		if err != nil {
			return "", errors.Wrap(errors.Wrap(err, path), "during interpolation")
		}
		params = append(params, staticVars)
	}

	res, err := NewTemplateResolver(config, params).Resolve(true)
	if err != nil {
		return "", err
	}

	return string(res), nil

}

type TemplateResolver struct {
	configPayload []byte
	params        []boshtemplate.Variables
}

func NewTemplateResolver(configPayload []byte, params []boshtemplate.Variables) TemplateResolver {
	return TemplateResolver{
		configPayload: configPayload,
		params:        params,
	}
}

func (resolver TemplateResolver) Resolve(expectAllKeys bool) ([]byte, error) {
	var err error

	resolver.configPayload, err = resolver.resolve(expectAllKeys)
	if err != nil {
		return nil, err
	}

	return resolver.configPayload, nil
}

func (resolver TemplateResolver) resolve(expectAllKeys bool) ([]byte, error) {
	tpl := boshtemplate.NewTemplate(resolver.configPayload)

	vars := []boshtemplate.Variables{}
	for i := len(resolver.params) - 1; i >= 0; i-- {
		vars = append(vars, resolver.params[i])
	}

	bytes, err := tpl.Evaluate(boshtemplate.NewMultiVars(vars), nil, boshtemplate.EvaluateOpts{ExpectAllKeys: expectAllKeys})

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

var (
	interpolationRegexBosh = regexp.MustCompile(`\(\((!?[-/\.\w\pL]+)\)\)`)
	interpolationRegexRuby = regexp.MustCompile(`<%= *([-_\w\pL]+) *%>`)
)

type TemplateType string

const (
	UnsupportedTemplateType = cterror.Error("unsupported template type")

	BoshTemplateType TemplateType = "bosh"
	RubyTemplateType              = "ruby"
)

func AllUniqueKeys(data string, templateType TemplateType) ([]string, error) {
	switch templateType {
	case BoshTemplateType:
		return AllUniqueKeysInBoshTemplate(data), nil
	case RubyTemplateType:
		return AllUniqueKeysInRubyTemplate(data), nil
	default:
		return []string{}, errors.Wrap(UnsupportedTemplateType, string(templateType))
	}
}

func AllUniqueKeysInRubyTemplate(data string) []string {
	names := make(map[string]struct{})

	for _, match := range interpolationRegexRuby.FindAllSubmatch([]byte(data), -1) {
		names[strings.TrimPrefix(string(match[1]), "!")] = struct{}{}
	}

	var res []string
	for key, _ := range names {
		res = append(res, key)
	}

	return res
}

func AllUniqueKeysInBoshTemplate(data string) []string {
	names := make(map[string]struct{})

	for _, match := range interpolationRegexBosh.FindAllSubmatch([]byte(data), -1) {
		names[strings.TrimPrefix(string(match[1]), "!")] = struct{}{}
	}

	var res []string
	for key, _ := range names {
		res = append(res, key)
	}

	return res
}
