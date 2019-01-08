package client

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/pkg/errors"
)

type OutputPair struct {
	Name string
	Path string
}

const (
	outputFlag         = "-o"
	inputsFromFlag     = "-j"
	taskConfigPathFlag = "-c"
)

func Execute(taskConfig, target, pipeline, job string, outputs []OutputPair) error {
	flycmd := NewFlyCmd().
		WithTarget(rc.TargetName(target)).
		WithSubCommand("execute")
	for _, output := range outputs {
		flycmd = flycmd.WithArg(outputFlag, fmt.Sprintf("%s=%s", output.Name, output.Path))
	}
	flycmd = flycmd.WithArg(inputsFromFlag, fmt.Sprintf("%s/%s", pipeline, job))

	tmpfile, err := ioutil.TempFile("", "fly-execute")
	if err != nil {
		return errors.Wrap(err, "fail to execute")
	}
	// clean up
	defer os.Remove(tmpfile.Name())

	err = io.WriteToFile(taskConfig, tmpfile.Name())
	if err != nil {
		return errors.Wrap(err, "fail to execute")
	}

	flycmd = flycmd.WithArg(taskConfigPathFlag, tmpfile.Name())

	err = flycmd.Run()
	if err != nil {
		return errors.Wrap(err, "fail to run command when executing")
	}
	return nil
}
