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

	TargetNotFoundError = cterror.Error("target not found")
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

func NewFlyClient() (*Client, error) {
	res := &Client{}
	path, err := io.BinaryPath(flyBinaryName)
	if err != nil {
		return res, errors.New(fmt.Sprintf("%s doesn't exist on your $PATH. install it before your use", flyBinaryName))
	}
	targets, err := rc.LoadTargets()
	if err != nil {
		return res, err
	}
	res.path = path
	res.targets = targets.Targets
	return res, nil
}

func LoadTargets() (Targets, error) {
	res := make(map[rc.TargetName]rc.TargetProps)
	targets, err := rc.LoadTargets()
	if err != nil {
		return Targets(res), err
	}
	res = targets.Targets
	return Targets(res), nil
}

func LoadTarget(name string) (rc.TargetProps, error) {
	targets, err := LoadTargets()
	if err != nil {
		return rc.TargetProps{}, err
	}
	t, ok := targets[rc.TargetName(name)]
	if !ok {
		return rc.TargetProps{}, TargetNotFoundError
	}
	return t, nil
}

func (c *Client) Targets() []rc.TargetName {
	var names []rc.TargetName
	for n, _ := range c.targets {
		names = append(names, n)
	}
	return names
}
