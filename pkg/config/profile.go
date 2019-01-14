package config

import (
	"os"
	"path"
	"path/filepath"

	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/maplain/control-tower/pkg/secret"
	"github.com/maplain/control-tower/templates"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	ProfileAlreadyExistError = cterror.Error("profile already exist")
	profileFolder            = ".control-tower-profile"
	profilesInfo             = ".profiles"

	EmptyProfileDeleteNameError = cterror.Error("name can not be empty")
	ProfileNotExistError        = cterror.Error("profile does not exist. please use `ct profile list` to check available profiles")

	DefaultEncryptionKey = "1234567891123456"
)

type Profile struct {
	Name         string
	TemplateKeys map[templates.TemplateType][]string
	Tags         Tags
	path         string
}

func (p Profile) IsTemplate() bool {
	if p.TemplateKeys == nil {
		return false
	}
	for _, keys := range p.TemplateKeys {
		if len(keys) != 0 {
			return true
		}
	}
	return false
}

type Profiles map[string]Profile

func (p Profiles) bytes() ([]byte, error) {
	data, err := yaml.Marshal(p)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (p Profiles) RemoveProfile(name string) {
	delete(p, name)
}

func (p Profiles) SaveProfile(profile Profile, overwrite bool, data string) error {
	path, err := getProfilePath(profile.Name)
	if err != nil {
		return err
	}
	_, ok := p[profile.Name]
	if ok {
		if !overwrite {
			return errors.Wrap(ProfileAlreadyExistError, profile.Name)
		}
	}

	profile.path = path
	p[profile.Name] = profile

	err = io.WriteToFile(data, path)
	if err != nil {
		return err
	}

	return nil
}

func (p Profiles) Save() error {
	path, err := getProfileControlInfoPath()
	if err != nil {
		return err
	}
	data, err := p.bytes()
	if err != nil {
		return err
	}
	err = io.WriteToFile(string(data), path)
	if err != nil {
		return err
	}
	return nil
}

func getProfileFolder() (string, error) {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return path.Join(home, profileFolder), nil
}

func getProfilePath(profilename string) (string, error) {
	folder, err := getProfileFolder()
	if err != nil {
		return "", err
	}
	return path.Join(folder, profilename), nil
}

func getProfileControlInfoPath() (string, error) {
	folder, err := getProfileFolder()
	if err != nil {
		return "", err
	}
	return path.Join(folder, profilesInfo), nil
}

func LoadProfileControlInfo() (Profiles, error) {
	res := Profiles(make(map[string]Profile))
	path, err := getProfileControlInfoPath()
	if err != nil {
		return res, err
	}

	if io.NotExist(path) {
		res, err = initializeProfileControlInfo()
		if err != nil {
			return res, err
		}
		err = res.Save()
		if err != nil {
			return res, err
		}
		return res, nil
	}

	data, err := io.ReadFromFile(path)
	if err != nil {
		return res, err
	}
	err = yaml.Unmarshal(data, &res)
	if err != nil {
		return res, err
	}
	return res, nil
}

func LoadProfile(name, key string) (string, error) {
	filepath, err := getProfilePath(name)
	if err != nil {
		return "", err
	}

	if io.NotExist(filepath) {
		return "", errors.Wrap(ProfileNotExistError, name)
	}

	ed, err := io.ReadFromFile(filepath)
	if err != nil {
		return "", err
	}

	d, err := secret.Decrypt(string(ed[:]), key)
	if err != nil {
		return "", err
	}
	return d, nil
}

func initializeProfileControlInfo() (Profiles, error) {
	res := Profiles(make(map[string]Profile))
	var files []string

	folder, err := getProfileFolder()
	if err != nil {
		return res, err
	}
	profilesInfo, err := getProfileControlInfoPath()
	if err != nil {
		return res, err
	}

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if path != folder && path != profilesInfo {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return res, err
	}
	for _, file := range files {
		profileName := filepath.Base(file)
		data, err := LoadProfile(profileName, DefaultEncryptionKey)
		if err != nil {
			return res, err
		}
		keys := templates.AllUniqueKeysInBoshTemplate(data)
		p := Profile{
			Name: profileName,
			path: file,
		}
		if len(keys) != 0 {
			p.TemplateKeys = make(map[templates.TemplateType][]string)
			p.TemplateKeys[templates.BoshTemplateType] = keys
		}
		res[profileName] = p
	}
	return res, nil
}
