[Control Tower](#Control-Tower)

[Installation](#Installation)

[Concepts](#Concepts)
 - [Profile](#Profile)
   - [Kubo related profiles setup](#Kubo-related-profiles-setup)
   - [RAAS related profiles setup](#RAAS-related-profiles-setup)
 - [Tags](#Tags)
 - [context](#context)
 - [Pipelines](#Pipelines)
   - [pks-nsx-t-release Pipeline](#pks-nsx-t-release-Pipeline)
   - [Built-in Pipelines](#Built-in-Pipelines)
    - [kubo pipeline](#kubo-pipeline)
    - [nsx-acceptance-tests Pipeline](#nsx-acceptance-tests-Pipeline)
    - [build-tile Pipeline](#build-tile-Pipeline)
    - [install-tile Pipeline](#install-tile-Pipeline)
    - [releng-acceptance-tests pipeline](#releng-acceptance-tests-pipeline)

[PKS Networking Workflow](#PKS-Networking-Workflow)

# Control Tower
Control Tower aims to provide a better abstraction over Concourse pipelines so that it's easier to manage and fly. For now, it's used internally in PKS networking team. Thus some commands are only available for a few built-in type of pipelines.


# Installation
```sh
# clone control-tower repo
git clone https://github.com/maplain/control-tower
# build ct binary
make build-ct
```
ct will be built under `bin`. You may want to optionally move it under your `$PATH`.

Or you can simply do a:
```sh
make build
```

# Concepts
## Profile
Profile is an encrypted key-value yaml file managed by ct. They are the source of truth when it comes to configure a templated pipeline.
You can create a profile from command line:
```sh
➜  control-tower git:(master) ✗ ct p c -n test --vars a=b,c=d
profile test is created successfully
➜  control-tower git:(master) ✗ ct p v -n test
a: b
c: d
```
Or from a key-value yaml file:
```sh
➜  control-tower git:(master) ✗ ct p c -n test-2 --var-file <(ct p v -n test)
profile test-2 is created successfully
➜  control-tower git:(master) ✗ ct p v -n test-2
a: b
c: d
```

It's ok to create profile with some fields templated by `(( ))`. For example:
```sh
➜  control-tower git:(master) ✗ ct p c -n test-3 --var-file template.yml --template
profile test-3 is created successfully
➜  control-tower git:(master) ✗ ct p v -n test-3
name: ((name))
➜  control-tower git:(master) ✗ cat template.yml
name: ((name))
```
Note: you need to use `--template` flag to inform ct that it's a templated profile.

When a template profile is used to create a pipeline, you'll be asked for the values by ct and one corresponding concrete profile will be optionally created.

Generally speaking, following practices are recommended:
* create normal profiles for static shared secrets
* create template profiles for dynamic configurations that vary from feature to feature, eg: branches, pool-name etc

This profiles are encrypted using a default key which might not suite your use case. To encrypt it using another key:
```sh
➜  control-tower git:(master) ✗ ct s g > secret
➜  control-tower git:(master) ✗ ct p c -n test-3 --vars a=b,c=d -k $(<secret)
profile test-3 is created successfully
➜  control-tower git:(master) ✗ ct p v -n test-3
2019/01/28 14:31:17 error: yaml: invalid leading UTF-8 octet
➜  control-tower git:(master) ✗ ct p v -n test-3 -k $(<secret)
a: b
c: d
```
### Kubo related profiles setup
To setup kubo-related profiles automatically:
```sh
make set-kubo
```
To update kubo related profiles forcefully:
```sh
make set-kubo-force
```


Kubo related profiles will be setup automatically for you. To check them out:
```sh
➜  control-tower git:(master) ✗ ct p l -t kubo
+---------------------------+------+------------+
|       PROFILE NAME        | TAGS | ISTEMPLATE |
+---------------------------+------+------------+
| common-secrets            | kubo | false      |
| deploy-kubo               | kubo | false      |
| kubo-fangyuanl            | kubo | false      |
| pks-nsx-t-release-secrets | kubo | false      |
+---------------------------+------+------------+
```
### RAAS related profiles setup
To setup raas-related profiles automatically:
```sh
make set-raas
```
To update kubo related profiles forcefully:
```sh
make set-raas-force
```
RAAS related profiles will be setup automatically for you. To check them out:
```sh
ct p l -t releng
+----------------------------+--------+------------+
|        PROFILE NAME        |  TAGS  | ISTEMPLATE |
+----------------------------+--------+------------+
| nsx-t-secrets              | releng | false      |
| pks-releng-write-locks-bot | releng | false      |
| raas-credentials           | releng | false      |
| vsphere-nsx-variables      | releng | false      |
+----------------------------+--------+------------+
ct p l -t releng-template
+----------------+-----------------+------------+
|  PROFILE NAME  |      TAGS       | ISTEMPLATE |
+----------------+-----------------+------------+
| raas-variables | releng-template | true       |
+----------------+-----------------+------------+
```
## Tags
You can group profiles by applying tags.
```sh
➜  control-tower git:(master) ✗ ct p t -n test -t test,example
tags of profile test are updated to example,test
```
To view profiles by tags:
```sh
➜  control-tower git:(master) ✗ ct p l -t example
+--------------+--------------+------------+
| PROFILE NAME |     TAGS     | ISTEMPLATE |
+--------------+--------------+------------+
| test         | example,test | false      |
+--------------+--------------+------------+
➜  control-tower git:(master) ✗ ct p l -t test
+--------------+--------------+------------+
| PROFILE NAME |     TAGS     | ISTEMPLATE |
+--------------+--------------+------------+
| test         | example,test | false      |
| test-2       | test         | false      |
+--------------+--------------+------------+
```

Tags are useful when:
* you want to deploy a pipeline with multiple profiles
* you want to separate similar configurations based on feature

## context
context is an alias for concourse target, team and pipeline. ct provides a rich set of utilities that are associated with one context.

Let's say you control following pipelines on concourse:
```sh
➜  control-tower git:(master) ✗ fly -t npks ps | grep kubo
fangyuanl-kubo-2                                            no   no
```
Create a context for pipeline `fangyuanl-kubo-2` by:
```sh
➜  control-tower git:(master) ✗ ct c c -n test --pipeline-name fangyuanl-kubo-2 -t npks
context test is created
use ct c v -n test to check details
```
Switch current context to it:
```sh
ct c set -n test
```
See the pipeline in browser:
```sh
ct o
```
Check out other contexts in browser:
```sh
ct o [another context name]
```
Check out jobs in current context:
```sh
ct f j
+------+---------------------------+
|  ID  |           NAME            |
+------+---------------------------+
| 9049 | claim-lock-kubo           |
| 9050 | run-precheck-release-lock |
| 9051 | run-precheck              |
| 9052 | deploy-kubo-release-lock  |
| 9053 | claim-lock-for-outputs    |
| 9054 | outputs                   |
| 9055 | outputs-release-lock      |
| 9056 | deploy-kubo               |
| 9057 | claim-lock-for-deletion   |
| 9058 | delete-kubo               |
| 9059 | delete-kubo-release-lock  |
+------+---------------------------+
```
Check out pipeline status in current context:
```sh
➜  control-tower git:(master) ✗ ct f s
+--------+------+------+-----------+-----------------+
|   ID   | TEAM | NAME |  STATUS   |       JOB       |
+--------+------+------+-----------+-----------------+
| 533232 | nsxt |    1 | succeeded | claim-lock-kubo |
| 539343 | nsxt |    2 | succeeded | claim-lock-kubo |
| 562675 | nsxt |    3 | succeeded | claim-lock-kubo |
| 562858 | nsxt |    4 | succeeded | claim-lock-kubo |
+--------+------+------+-----------+-----------------+
+--------+------+------+-----------+---------------------------+
|   ID   | TEAM | NAME |  STATUS   |            JOB            |
+--------+------+------+-----------+---------------------------+
| 539080 | nsxt |    1 | succeeded | run-precheck-release-lock |
+--------+------+------+-----------+---------------------------+
+--------+------+------+-----------+--------------+
|   ID   | TEAM | NAME |  STATUS   |     JOB      |
+--------+------+------+-----------+--------------+
| 573958 | nsxt |   23 | succeeded | run-precheck |
| 574034 | nsxt |   24 | succeeded | run-precheck |
| 574161 | nsxt |   25 | failed    | run-precheck |
| 574171 | nsxt |   26 | failed    | run-precheck |
| 574604 | nsxt |   27 | failed    | run-precheck |
+--------+------+------+-----------+--------------+
+--------+------+------+-----------+--------------------------+
|   ID   | TEAM | NAME |  STATUS   |           JOB            |
+--------+------+------+-----------+--------------------------+
| 553764 | nsxt |    1 | succeeded | deploy-kubo-release-lock |
| 562845 | nsxt |    2 | succeeded | deploy-kubo-release-lock |
| 563411 | nsxt |    3 | succeeded | deploy-kubo-release-lock |
+--------+------+------+-----------+--------------------------+
+--------+------+------+-----------+-------------+
|   ID   | TEAM | NAME |  STATUS   |     JOB     |
+--------+------+------+-----------+-------------+
| 563232 | nsxt |   13 | succeeded | deploy-kubo |
| 567257 | nsxt |   14 | succeeded | deploy-kubo |
| 567721 | nsxt |   15 | failed    | deploy-kubo |
| 573968 | nsxt |   16 | errored   | deploy-kubo |
| 574036 | nsxt |   17 | errored   | deploy-kubo |
+--------+------+------+-----------+-------------+
```
Check out logs for a specific build:
```sh
➜  control-tower git:(master) ✗ ct f logs -i 533232
acquiring lock on: fangyuanl
Cloning into '/tmp/build/get'...
5478c9162c fangyuanl-kubo-2/claim-lock-kubo build 1 claiming: nsx2
```

You can also provide customized `outputs` for built-in type pipelines:
```sh
➜  control-tower git:(master) ✗ ct c set kubo
current context is set to kubo
➜  control-tower git:(master) ✗ ct c v
target: npks
team: nsxt
pipeline: fangyuanl-kubo-2
type: kubo
inuse: true
➜  control-tower git:(master) ✗ ct f o
build-id: "567257"
kubeconfig-path: fangyuanl/kubeconfig-0.270.0.tgz
lock-name: nsx3
```

The `outputs` for type `kubo` pipeline is defined to be a key-value yaml file so that it can be directly piped into another command's input. See examples below.

## Pipelines
To deploy a pipeline, basically there are three kinds of information needed:
1. pipeline yaml file which can be specified by either `--template` or `--template-type`. `--template` accepts a filepath, whereas `--template-type` accepts a list of reserved static types. See more details below;
2.(optional) fly target and pipeline name. If not provided, values in current context are used;
3.(optional) profiles that used to populate pipeline template yaml file. Three flags are available: `--profile-name`, `--profile-path` and `--profile-tag`.

To login the context:
```sh
ct f login
```

### pks-nsx-t-release Pipeline
Deploy the pipeline:
```sh
pushd $GOPATH/src/gitlab.eng.vmware.com/PKS/pks-nsx-t-release
  # ci/pks-nsx-t-release.yml is templated using erb semantics
  ct d --profile-tag kubo -m <(erb ci/pks-nsx-t-release.yml) -n [pipeline name]
popd
```

### Built-in Pipelines
#### kubo pipeline
Deploy kubo pipeline
```sh
ct d --profile-tag=kubo --template-type kubo -n kubo --target npks
```
#### nsx-acceptance-tests Pipeline
If your current context is on a `kubo` type pipeline and it has run through successfully, then you can deploy a `nsx-acceptance-tests` pipeline with artifacts from that `kubo` pipeline.
```sh
➜  control-tower git:(master) ✗ ct c c -n ntests --pipeline-name nsx-acceptance-tests --target npks
context ntests is created
use ct c v -n ntests to check details
➜  control-tower git:(master) ✗ ct c set ntests
current context is set to ntests

➜  control-tower git:(master) ✗ ct d --profile-tag=kubo --template-type nsx-acceptance-tests --profile-path=<(ct f outputs) 
```
In above example, we use the `outputs` from a `kubo` type pipeline to configure a `nsx-acceptance-tests` type pipeline so that the latter will use the exact versions of `kubeconfig` and `lock` generated by the first one.

To check all jobs defined for `nsx-acceptance-tests` type of pipeline:
```sh
➜  control-tower git:(master) ✗ ct f j
+------+------------------------------------+
|  ID  |                NAME                |
+------+------------------------------------+
| 9427 | claim-lock                         |
| 9428 | run-release-tests-release-lock     |
| 9429 | run-release-tests                  |
| 9430 | run-conformance-tests-release-lock |
| 9431 | run-conformance-tests-delete-kubo  |
| 9432 | run-conformance-tests              |
| 9433 | run-release-tests-delete-kubo      |
+------+------------------------------------+
```
To unpause the pipeline:
```sh
➜  control-tower git:(master) ✗ ct f p -u
pipeline fangyuanl-kubo-2 is unpaused.
```
To trigger the first job:
```sh
➜  control-tower git:(master) ✗ ct f trigger -j claim-lock
started nsx-acceptance-tests/claim-lock #1
```
To pause the pipeline:
```sh
➜  control-tower git:(master) ✗ ct f p
pipeline nsx-acceptance-tests is paused.
```

To get pipeline configuration yaml:
```sh
ct f c
```

To get job configuration:
```sh
ct f c -j [job-name]
```

To delete the pipeline:
```
➜  control-tower git:(master) ✗ ct f d
pipeline nsx-acceptance-tests is deleted. now ntests is a dangling context%
➜  control-tower git:(master) ✗ ct c d -n ntests
context ntests is deleted successfully
```
#### build-tile Pipeline
This pipeline will build a PKS tile based on your configuration
```sh
ct d --profile-tag=releng --profile-tag=releng-template --template-type build-tile -n fangyuanl-kubo-2 --target npks
```
#### install-tile Pipeline
This pipeline will install a PKS tile to the specified Nimbus testbed
```sh
ct d --profile-tag=releng --profile-tag=releng-template --template-type install-tile -n fangyuanl-kubo-2 --target npks
```
#### releng-acceptance-tests pipeline
This pipeline will deploy a k8s cluster in given PKS deployment and run acceptance-tests
```sh
ct d --profile-path <(echo "releng-tests-lock-name: nsx1") -p common-secrets --profile-tag releng --profile-tag nodes_dns --template-type releng-acceptance-tests
```

# PKS Networking Workflow
check out [doc](https://github.com/maplain/control-tower/blob/master/FEATURE-DEVELOPMENT.md)
