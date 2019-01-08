package templates

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/concourse/atc"
	"github.com/maplain/control-tower/pkg/concourseclient"
	client "github.com/maplain/control-tower/pkg/flyclient"
	"github.com/pkg/errors"
)

const (
	ArtifactsOutputVolume = "artifacts"
	ArtifactsJobName      = "outputs"

	JobStatusArgKey = "status"

	deployKuboArtifactsTaskConfig = `
platform: linux
image_resource:
  type: docker-image
  source:
    repository: gcr.io/eminent-nation-87317/pks-ci
    tag: stable
run:
  path: pks-concourse/scripts/deploy-kubo-outputs.sh

inputs:
- name: pks-lock
- name: kubeconfig
- name: pks-concourse
outputs:
- name: artifacts
`
)

type Args map[string]interface{}

func (a Args) Add(key string, value interface{}) {
	a[key] = value
}

type PipelineFetchOutputFunc func(team, pipeline string, cli concourseclient.ConcourseClient, args Args) (PipelineOutput, error)
type PipelineGetArtifactsFunc func(target, pipeline string) (string, error)
type PipelineOutput map[string]string

var DeployKuboPipelineGetArtifactsFunc = PipelineGetArtifactsFunc(deployKuboPipelineGetArtifacts)
var DeployKuboPipelineFetchOutputFunc = PipelineFetchOutputFunc(deployKuboPipelineOutput)
var NsxAcceptanceTestsPipelineFetchOutputFunc = PipelineFetchOutputFunc(nsxAcceptanceTestsPipelineOutput)

func deployKuboPipelineGetArtifacts(target, pipeline string) (string, error) {
	dir, err := ioutil.TempDir("", "fly")
	if err != nil {
		return "", errors.Wrap(err, "can not get artifacts for deploy-kubo type pipeline")
	}
	outputs := []client.OutputPair{
		client.OutputPair{
			Name: ArtifactsOutputVolume,
			Path: dir,
		},
	}
	err = client.Execute(deployKuboArtifactsTaskConfig, target, pipeline, ArtifactsJobName, outputs)
	if err != nil {
		return "", errors.WithMessage(errors.Wrap(err, "can not get artifacts for deploy-kubo type pipeline"), pipeline)
	}
	return dir, nil
}

// deploy-kubo template specific logic
func deployKuboPipelineOutput(team, pipeline string, cli concourseclient.ConcourseClient, args Args) (PipelineOutput, error) {
	res := make(map[string]string)

	status := string(atc.StatusSucceeded)
	for k, v := range args {
		if k == "status" {
			vs, ok := v.(string)
			if ok {
				status = vs
			}
		}
	}

	id, err := cli.LatestJobBuildIDOnStatus(team, pipeline, "deploy-kubo", status)
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

// nsx-acceptance-tests template specific logic
func nsxAcceptanceTestsPipelineOutput(team, pipeline string, cli concourseclient.ConcourseClient, args Args) (PipelineOutput, error) {
	res := make(map[string]string)
	build, err := cli.LatestJobBuild(team, pipeline, "run-release-tests")
	if err != nil {
		if err != concourseclient.BuildNotFoundError {
			return PipelineOutput(res), err
		}
	} else {
		res["run-release-tests-status"] = build.Status
		res["run-release-tests-id"] = strconv.Itoa(build.ID)
		res["run-release-tests-name"] = build.Name
	}

	build, err = cli.LatestJobBuild(team, pipeline, "run-conformance-tests")
	if err != nil {
		if err != concourseclient.BuildNotFoundError {
			return PipelineOutput(res), err
		}
	} else {
		res["run-conformance-tests-status"] = build.Status
		res["run-conformance-tests-id"] = strconv.Itoa(build.ID)
		res["run-conformance-tests-name"] = build.Name
	}
	return PipelineOutput(res), nil
}
