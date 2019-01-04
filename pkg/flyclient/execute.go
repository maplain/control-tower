package client

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/concourse/atc"
	"github.com/concourse/fly/config"
	"github.com/concourse/fly/eventstream"
	"github.com/concourse/fly/rc"
	"github.com/concourse/fly/ui"
	"github.com/concourse/go-concourse/concourse"
	"github.com/maplain/control-tower/pkg/flyclient/executehelpers"
	"github.com/maplain/control-tower/pkg/flyclient/flaghelpers"
)

func Execute(taskConfig, target, pipeline, job string, outputs []flaghelpers.OutputPairFlag) error {
	cmd := ExecuteCommand{
		TaskConfig: taskConfig,
		InputsFrom: flaghelpers.JobFlag{pipeline, job},
		Outputs:    outputs,
	}
	return cmd.Execute(target, []string{})
}

type ExecuteCommand struct {
	TaskConfig     string
	Privileged     bool
	IncludeIgnored bool
	Inputs         []flaghelpers.InputPairFlag
	InputsFrom     flaghelpers.JobFlag
	Outputs        []flaghelpers.OutputPairFlag
	Tags           []string
}

func (command *ExecuteCommand) Execute(t string, args []string) error {
	target, err := rc.LoadTarget(rc.TargetName(t), false)
	if err != nil {
		return err
	}

	err = target.Validate()
	if err != nil {
		return err
	}

	taskConfigFile := command.TaskConfig
	includeIgnored := command.IncludeIgnored

	taskConfig, err := config.LoadTaskConfig(taskConfigFile, args)
	if err != nil {
		return err
	}

	client := target.Client()

	fact := atc.NewPlanFactory(time.Now().Unix())

	inputs, err := executehelpers.DetermineInputs(
		fact,
		target.Team(),
		taskConfig.Inputs,
		command.Inputs,
		command.InputsFrom,
	)
	if err != nil {
		return err
	}

	outputs, err := executehelpers.DetermineOutputs(
		fact,
		taskConfig.Outputs,
		command.Outputs,
	)
	if err != nil {
		return err
	}

	plan, err := executehelpers.CreateBuildPlan(
		fact,
		target,
		command.Privileged,
		inputs,
		outputs,
		taskConfig,
		command.Tags,
	)

	if err != nil {
		return err
	}

	clientURL, err := url.Parse(client.URL())
	if err != nil {
		return err
	}

	var build atc.Build
	var buildURL *url.URL

	if command.InputsFrom.PipelineName != "" {
		build, err = target.Team().CreatePipelineBuild(command.InputsFrom.PipelineName, plan)
		if err != nil {
			return err
		}

		buildURL, err = url.Parse(fmt.Sprintf("/teams/%s/pipelines/%s/builds/%s", build.TeamName, build.PipelineName, build.Name))
		if err != nil {
			return err
		}

	} else {
		build, err = target.Team().CreateBuild(plan)
		if err != nil {
			return err
		}

		buildURL, err = url.Parse(fmt.Sprintf("/builds/%d", build.ID))
		if err != nil {
			return err
		}
	}

	fmt.Printf("executing build %d at %s \n", build.ID, clientURL.ResolveReference(buildURL))

	terminate := make(chan os.Signal, 1)

	go abortOnSignal(client, terminate, build)

	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	inputChan := make(chan interface{})
	go func() {
		for _, i := range inputs {
			if i.Path != "" {
				executehelpers.Upload(client, build.ID, i, includeIgnored)
			}
		}
		close(inputChan)
	}()

	var outputChans []chan (interface{})
	if len(outputs) > 0 {
		for i, output := range outputs {
			outputChans = append(outputChans, make(chan interface{}, 1))
			go func(o executehelpers.Output, outputChan chan<- interface{}) {
				if o.Path != "" {
					executehelpers.Download(client, build.ID, o)
				}

				close(outputChan)
			}(output, outputChans[i])
		}
	}

	eventSource, err := client.BuildEvents(fmt.Sprintf("%d", build.ID))
	if err != nil {
		return err
	}

	exitCode := eventstream.Render(os.Stdout, eventSource)
	eventSource.Close()

	<-inputChan

	if len(outputs) > 0 {
		for _, outputChan := range outputChans {
			<-outputChan
		}
	}

	os.Exit(exitCode)

	return nil
}

func abortOnSignal(
	client concourse.Client,
	terminate <-chan os.Signal,
	build atc.Build,
) {
	<-terminate

	fmt.Fprintf(ui.Stderr, "\naborting...\n")

	err := client.AbortBuild(strconv.Itoa(build.ID))
	if err != nil {
		fmt.Fprintln(ui.Stderr, "failed to abort:", err)
		return
	}

	// if told to terminate again, exit immediately
	<-terminate
	fmt.Fprintln(ui.Stderr, "exiting immediately")
	os.Exit(2)
}
