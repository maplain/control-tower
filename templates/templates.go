package templates

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/concourse/atc"
	"github.com/maplain/control-tower/pkg/concourseclient"
)

type PipelineFetchOutputFunc func(team, pipeline string, cli concourseclient.ConcourseClient) (PipelineOutput, error)
type PipelineOutput map[string]string

var DeployKuboPipelineFetchOutputFunc = PipelineFetchOutputFunc(deployKuboPipelineOutput)

// deploy-kubo template specific logic
func deployKuboPipelineOutput(team, pipeline string, cli concourseclient.ConcourseClient) (PipelineOutput, error) {
	res := make(map[string]string)

	id, err := cli.LatestJobBuildIDOnStatus(team, pipeline, "deploy-kubo", string(atc.StatusSucceeded))
	if err != nil {
		return PipelineOutput(res), err
	}
	res["build-id"] = strconv.Itoa(id)

	resources, found, err := cli.Client().BuildResources(id)
	if err != nil {
		return PipelineOutput(res), err
	}

	if !found {
		return PipelineOutput(res), errors.New(fmt.Sprintf("no resource found for job deploy-kubo in team %s pipeline %s", team, pipeline))
	}
	for _, input := range resources.Inputs {
		if input.Resource == "pks-lock" {
			for _, metadata := range input.Metadata {
				if metadata.Name == "lock_name" {
					res["lock-name"] = metadata.Value
				}
			}
		}
	}

	for _, output := range resources.Outputs {
		if output.Resource == "kubeconfig" {
			for key, val := range output.Version {
				if key == "path" {
					res["kubeconfig-path"] = val
				}
			}
		}
	}

	return PipelineOutput(res), nil
}
