package config

import (
	"path"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	ConfigFilename = ".control-tower"
	ProfileFolder  = ".control-tower-profile"
)

func GetProfilePath(profilename string) (string, error) {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return path.Join(home, ProfileFolder, profilename), nil
}
