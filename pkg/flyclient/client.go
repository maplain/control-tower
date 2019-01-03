package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/concourse/fly/rc"
	cterror "github.com/maplain/control-tower/pkg/error"
	cio "github.com/maplain/control-tower/pkg/io"
	"github.com/skratchdot/open-golang/open"
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
	flags  []string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

type Client struct {
	path    string
	targets Targets
}

func NewFlyCmd() *Cmd {
	res := &Cmd{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	path, err := cio.BinaryPath(flyBinaryName)
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

func (c *Cmd) WithFlag(flag string) *Cmd {
	c.flags = append(c.flags, flag)
	return c
}
func (c *Cmd) WithStdin(reader io.Reader) *Cmd {
	c.stdin = reader
	return c
}

func (c *Cmd) WithStdout(writer io.Writer) *Cmd {
	c.stdout = writer
	return c
}

func (c *Cmd) WithStderr(writer io.Writer) *Cmd {
	c.stderr = writer
	return c
}

func (c *Cmd) Cmd() *exec.Cmd {
	args := []string{}
	if c.target != "" {
		args = append(args, "--target", string(c.target))
	}
	if c.subcmd != "" {
		args = append(args, c.subcmd)
	}
	args = append(args, c.flags...)
	args = append(args, c.args...)
	cmd := exec.Command(flyBinaryName, args...)
	return cmd
}

func (c *Cmd) Run() error {
	cmd := c.Cmd()
	cmd.Stdin = c.stdin
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func NewFlyClient() (*Client, error) {
	res := &Client{}
	path, err := cio.BinaryPath(flyBinaryName)
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

func SimpleLogin(target string) error {
	return NewFlyCmd().
		WithTarget(rc.TargetName(target)).
		WithSubCommand("login").
		WithFlag("-b").
		Run()
}

func Login(target string) error {
	flycmd := NewFlyCmd().
		WithTarget(rc.TargetName(target)).
		WithSubCommand("login").
		Cmd()

	stdout, err := flycmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := flycmd.StderrPipe()
	if err != nil {
		return err
	}
	flycmd.Stdin = os.Stdin

	err = flycmd.Start()
	if err != nil {
		return err
	}

	errc := make(chan error)
	// stdout will have the webhook
	go func() {
		stdoutReader := bufio.NewReader(stdout)
		line, err := stdoutReader.ReadString('\n')
		for err == nil {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "https://") {
				errc <- open.Run(line)
				return
			}
			line, err = stdoutReader.ReadString('\n')
		}
		if err == io.EOF {
			errc <- nil
		} else {
			errc <- err
		}
		return
	}()
	// stderr will have the warning for sync, if there is any
	go func() {
		stderrReader := bufio.NewReader(stderr)
		line, err := stderrReader.ReadString('\n')
		for err == nil {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "out of sync") {
				errc <- NewFlyCmd().
					WithTarget(rc.TargetName(target)).
					WithSubCommand("sync").
					Run()
				return
			}
			line, err = stderrReader.ReadString('\n')
		}
		if err == io.EOF {
			errc <- nil
		} else {
			errc <- err
		}
		return
	}()
	var res string
	for i := 0; i < 2; i++ {
		err = <-errc
		if err != nil {
			res = res + err.Error()
		}
	}
	close(errc)
	err = flycmd.Wait()
	if err != nil {
		res = res + err.Error()
	}
	if res == "" {
		return nil
	}
	return errors.New(res)
}
