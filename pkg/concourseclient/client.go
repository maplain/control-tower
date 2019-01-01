package concourseclient

import (
	"github.com/concourse/fly/rc"
	"github.com/concourse/go-concourse/concourse"

	cterror "github.com/maplain/control-tower/pkg/error"
)

const (
	clientTracing       = false
	TargetNotFoundError = cterror.Error("target is not found")
	BuildNotFoundError  = cterror.Error("build not found")
	QueryBuildLimit     = 20
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
		builds, pagination, ifpagination, err = t.JobBuilds(pipeline, job, *pagination.Next)
		if err != nil {
			return -1, err
		}
	}
	return -1, BuildNotFoundError
}
