package client

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

type Targets struct {
	Targets map[string]Target `yaml:"targets"`
}
