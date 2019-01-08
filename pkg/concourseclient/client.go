package concourseclient

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/concourse/atc"
	"github.com/concourse/atc/event"
	"github.com/concourse/fly/rc"
	"github.com/concourse/go-concourse/concourse"

	cterror "github.com/maplain/control-tower/pkg/error"
	client "github.com/maplain/control-tower/pkg/flyclient"
)

const (
	clientTracing   = false
	QueryBuildLimit = 20

	TargetNotFoundError  = cterror.Error("target is not found")
	BuildNotFoundError   = cterror.Error("build not found")
	NotEventLogTypeError = cterror.Error("not a log type event")
	NoAvailableJobsError = cterror.Error("no available jobs")
)

// ConcourseClient interface is a combination of original go-concourse Client
// and Control-Tower Concourse Client which implements other methods as a complementary
// interface
type ConcourseClient interface {
	CTConcourseClient
	rc.Target
}

type CTConcourseClient interface {
	LatestJobBuildIDOnStatus(team, pipeline, job, status string) (int, error)
	LatestJobBuild(team, pipeline, job string) (atc.Build, error)
	ReadBuildLog(id string, writer io.Writer) error
}

func NewConcourseClient(target rc.TargetName) (*oldCClient, error) {
	t, err := rc.LoadTarget(target, clientTracing)
	res := &oldCClient{}
	res.Target = t
	return res, err
}

type oldCClient struct {
	rc.Target
}

func ConvertEventToEnvelope(e atc.Event) (event.Envelope, error) {
	var envelope event.Envelope
	data, err := json.Marshal(event.Message{e})
	if err != nil {
		return envelope, err
	}
	err = json.Unmarshal(data, &envelope)
	return envelope, err
}

func GetEventLog(envelope event.Envelope) (event.Log, error) {
	var eventLog event.Log
	if envelope.Event == event.EventTypeLog {
		data, err := json.Marshal(envelope.Data)
		if err != nil {
			return eventLog, err
		}
		err = json.Unmarshal(data, &eventLog)
		return eventLog, err
	}
	return event.Log{}, NotEventLogTypeError
}

func (c *oldCClient) ReadBuildLog(id string, writer io.Writer) error {
	events, err := c.Client().BuildEvents(id)
	if err != nil {
		return err
	}
	e, err := events.NextEvent()
	for err == nil {
		if e.EventType() == event.EventTypeLog {
			envelope, err := ConvertEventToEnvelope(e)
			if err != nil {
				return err
			}
			eventLog, err := GetEventLog(envelope)
			if err != nil {
				return err
			}
			_, err = io.WriteString(writer, eventLog.Payload)
			if err != nil {
				return err
			}
		}
		e, err = events.NextEvent()
	}
	if err != io.EOF {
		return err
	}
	return nil
}

func (c *oldCClient) LatestJobBuild(team, pipeline, job string) (atc.Build, error) {
	t := c.Client().Team(team)
	builds, _, _, err := t.JobBuilds(pipeline, job, concourse.Page{Limit: QueryBuildLimit})
	if err != nil {
		return atc.Build{}, err
	}
	if len(builds) > 0 {
		return builds[0], nil
	}
	return atc.Build{}, BuildNotFoundError
}

func (c *oldCClient) LatestJobBuildIDOnStatus(team, pipeline, job, status string) (int, error) {
	t := c.Client().Team(team)
	builds, pagination, ifpagination, err := t.JobBuilds(pipeline, job, concourse.Page{Limit: QueryBuildLimit})
	if err != nil {
		return -1, err
	}
	for ifpagination {
		for _, build := range builds {
			if build.Status == status {
				return build.ID, nil
			}
		}
		if pagination.Next == nil {
			return -1, NoAvailableJobsError
		}
		builds, pagination, ifpagination, err = t.JobBuilds(pipeline, job, *pagination.Next)
		if err != nil {
			return -1, err
		}
	}
	return -1, BuildNotFoundError
}

func GetPipelineURL(target, pipeline string) (string, error) {
	t, err := client.LoadTarget(target)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/teams/%s/pipelines/%s", t.API, t.TeamName, pipeline), nil
}
