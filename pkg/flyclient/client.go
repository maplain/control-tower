package client

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/concourse/fly/rc"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
)

const (
	flyBinaryName        = "fly"
	flyConfigurationFile = ".flyrc"
)

type Cmd struct {
	path   string
	target rc.TargetName
	subcmd string
	args   []string
}

type Client struct {
	path    string
	targets Targets
}

func NewFlyCmd() *Cmd {
	res := &Cmd{}
	path, err := io.BinaryPath(flyBinaryName)
	if err != nil {
		cterror.Check(errors.New(fmt.Sprintf("%s doesn't exist on your $PATH. install it before your use", flyBinaryName)))
	}
	res.path = path
	return res
}

func (c *Cmd) WithTarget(target rc.TargetName) *Cmd {
	c.target = target
	return c
}

func (c *Cmd) WithSubCommand(cmd string) *Cmd {
	c.subcmd = cmd
	return c
}

func (c *Cmd) WithArg(arg, value string) *Cmd {
	c.args = append(c.args, arg, value)
	return c
}

func (c *Cmd) Run() error {
	args := []string{}
	if c.target != "" {
		args = append(args, "--target", string(c.target))
	}
	if c.subcmd != "" {
		args = append(args, c.subcmd)
	}
	args = append(args, c.args...)
	cmd := exec.Command(flyBinaryName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func NewFlyClient() *Client {
	res := &Client{}
	path, err := io.BinaryPath(flyBinaryName)
	if err != nil {
		cterror.Check(errors.New(fmt.Sprintf("%s doesn't exist on your $PATH. install it before your use", flyBinaryName)))
	}
	targets, err := rc.LoadTargets()
	if err != nil {
		cterror.Check(err)
	}
	res.path = path
	res.targets = targets.Targets
	return res
}

// func InitializeTargetsFromCfg() (Targets, error) {
// 	// Find home directory.
// 	home, err := homedir.Dir()
// 	if err != nil {
// 		return Targets{}, err
// 	}
//
// 	flyCfg := path.Join(home, flyConfigurationFile)
// 	if !io.NotExist(flyCfg) {
// 		d, err := io.ReadFromFile(flyCfg)
// 		cterror.Check(err)
//
// 		targets := Targets{}
// 		err = yaml.Unmarshal(d, &targets)
// 		cterror.Check(err)
//
// 		return targets, nil
// 	}
// 	return Targets{}, errors.New(fmt.Sprintf("could not find %s", flyCfg))
// }
//

func (c *Client) Targets() []rc.TargetName {
	var names []rc.TargetName
	for n, _ := range c.targets {
		names = append(names, n)
	}
	return names
}
