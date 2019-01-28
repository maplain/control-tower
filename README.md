[Concepts](#concepts)
- [Profile](#profile)
  -  [Tags](#tags)
- [Context](#context)
- [Kubo Related Pipelines](#kubo-related-pipelines)
- [RAAS Related Pipelines](#raas-related-pipelines)

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

### Tags
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

## Kubo Related Pipelines
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
Deploy `pks-nsx-t-release` pipeline:
```sh
pushd $GOPATH/src/gitlab.eng.vmware.com/PKS/pks-nsx-t-release
  ct d --profile-tag kubo -m <(erb ci/pks-nsx-t-release.yml) -n [pipeline name]
popd
```

## RAAS Related Pipelines
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
