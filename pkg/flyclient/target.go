package client

import "github.com/concourse/fly/rc"

type Target struct {
	API      string `yaml:"api"`
	Team     string `yaml:"team"`
	Insecure string `yaml:"insecure"`
	Token    Token  `yaml:"token"`
}

type Token struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

type Targets map[rc.TargetName]rc.TargetProps
