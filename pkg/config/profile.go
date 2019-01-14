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
	OverwriteExistingProfile = true
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
	Path         string
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

type NamedProfiles map[string]Profile
type TagedProfiles map[string]io.StringSet

type Profiles struct {
	NamedProfiles
	// saves mapping between tag and profile name
	TagedProfiles
}

func NewProfiles() Profiles {
	return Profiles{
		TagedProfiles: make(map[string]io.StringSet),
		NamedProfiles: make(map[string]Profile),
	}
}

func (p Profiles) GetProfileInfos() []Profile {
	res := []Profile{}
	for _, profile := range p.NamedProfiles {
		res = append(res, profile)
	}
	return res
}

func (p Profiles) GetProfilesByTag(tag string) []Profile {
	var res []Profile
	names := p.TagedProfiles[tag]
	for n, _ := range names {
		res = append(res, p.NamedProfiles[n])
	}
	return res
}

func (p Profiles) GetProfileInfoByName(name string) (Profile, error) {
	profile, found := p.NamedProfiles[name]
	if !found {
		return profile, errors.Wrap(ProfileNotExistError, name)
	}
	return profile, nil
}

func (p Profiles) bytes() ([]byte, error) {
	data, err := yaml.Marshal(p)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (p Profiles) RemoveProfileByName(name string) {
	// remove profile content from disk
	profile, found := p.NamedProfiles[name]
	if found {
		os.Remove(profile.Path)
	}

	// remove profile info from profiles
	delete(p.NamedProfiles, name)
	// for each tag, remove profile from corresponding list
	for t, _ := range profile.Tags.StringSet {
		p.TagedProfiles[t].Remove(name)
	}
}

// LoadProfile reads in profile content by name and a decryption key
func (p Profiles) LoadProfileByName(name, key string) (string, error) {
	profile, found := p.NamedProfiles[name]
	if !found {
		return "", errors.Wrap(ProfileNotExistError, name)
	}

	d, err := readProfileFile(profile.Path, key)
	return d, err
}

// readProfileFile reads in profile content by profile path and a decryption key
func readProfileFile(path, key string) (string, error) {
	ed, err := io.ReadFromFile(path)
	if err != nil {
		return "", err
	}

	d, err := secret.Decrypt(string(ed[:]), key)
	if err != nil {
		return "", err
	}
	return d, nil
}

// SaveProfileInfo updates profile info in profiles object
func (p Profiles) SaveProfileInfo(profile Profile, overwrite bool) error {
	var path string
	var err error

	_, ok := p.NamedProfiles[profile.Name]
	if ok {
		if !overwrite {
			return errors.Wrap(ProfileAlreadyExistError, profile.Name)
		}
	}

	// by default use path in profile struct
	path = profile.Path
	if path == "" {
		// if path is empty, fall back to default path
		path, err = getProfilePath(profile.Name)
		if err != nil {
			return err
		}
	}

	// overwrite profile info in profiles struct
	profile.Path = path
	p.NamedProfiles[profile.Name] = profile

	for t, _ := range profile.Tags.StringSet {
		set := p.TagedProfiles[t]
		if set == nil {
			p.TagedProfiles[t] = io.NewStringSet()
		}
		p.TagedProfiles[t].Add(profile.Name)
	}

	return nil
}

// SaveProfile adds one profile under profiles' control
// it adds profile info and saves its content on disk as well
func (p Profiles) SaveProfile(profile Profile, overwrite bool, data string) error {
	err := p.SaveProfileInfo(profile, overwrite)
	if err != nil {
		return err
	}

	// persist profile on disk
	err = io.WriteToFile(data, p.NamedProfiles[profile.Name].Path)
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
	res := NewProfiles()
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

func initializeProfileControlInfo() (Profiles, error) {
	res := NewProfiles()
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

		data, err := readProfileFile(file, DefaultEncryptionKey)
		if err != nil {
			return res, err
		}

		p := Profile{
			Name: profileName,
			Path: file,
		}

		keys := templates.AllUniqueKeysInBoshTemplate(data)
		if len(keys) != 0 {
			p.TemplateKeys = make(map[templates.TemplateType][]string)
			p.TemplateKeys[templates.BoshTemplateType] = keys
		}

		res.NamedProfiles[profileName] = p
	}
	return res, nil
}
