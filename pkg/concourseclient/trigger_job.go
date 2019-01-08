package concourseclient

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/concourse/fly/eventstream"
	"github.com/concourse/fly/rc"
	"github.com/concourse/fly/ui"
)

func TriggerJob(target, pipelineName, jobName string, watch bool) error {
	t, err := rc.LoadTarget(rc.TargetName(target), clientTracing)
	if err != nil {
		return err
	}

	err = t.Validate()
	if err != nil {
		return err
	}

	build, err := t.Team().CreateJobBuild(pipelineName, jobName)
	if err != nil {
		return err
	}
	fmt.Printf("started %s/%s #%s\n", pipelineName, jobName, build.Name)

	if watch {
		terminate := make(chan os.Signal, 1)

		go func(terminate <-chan os.Signal) {
			<-terminate
			fmt.Fprintf(ui.Stderr, "\ndetached, build is still running...\n")
			fmt.Fprintf(ui.Stderr, "re-attach to it with:\n\n")
			fmt.Fprintf(ui.Stderr, "    "+ui.Embolden(fmt.Sprintf("fly -t %s watch -j %s/%s -b %s\n\n", t, pipelineName, jobName, build.Name)))
			os.Exit(2)
		}(terminate)

		signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

		fmt.Println("")
		eventSource, err := t.Client().BuildEvents(fmt.Sprintf("%d", build.ID))
		if err != nil {
			return err
		}

		exitCode := eventstream.Render(os.Stdout, eventSource)

		eventSource.Close()

		os.Exit(exitCode)
	}

	return nil

}
