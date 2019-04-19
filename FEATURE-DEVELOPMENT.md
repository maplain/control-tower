- [Background](#Background)
- [Step1 Development](#Step1-Development)
- [Step2 Create pipeline to build dev release](#Step2-Create-pipeline-to-build-dev-release)
  - [Create profile for release storage options](#Create-profile-for-release-storage-options)
  - [Create pipeline](#Create-pipeline)
- [Step3 Bump release version](#Step3-Bump-release-version)
  - [p-pks-integrations](#p-pks-integrations)
  - [pks-releng-ci](#pks-releng-ci)
- [Step4 Raas Pipelines](#Step4-Raas-Pipelines)
  - [Create profiles for automation credentials](#Create-profiles-for-automation-credentials)
  - [create pipeline](#create-pipeline)
- [Step5 Install Dev PKS](#Step5-Install-Dev-PKS)
- [Summary](#Summary)

PKS networking team is always striving to automate as much as possible to make daily life easier. This document describes the workflow of feature development in our team.

### Background
Pipelines managing tile lifecycle like: tile creation, configuration, deployment etc are managed by team `Release Engineering` in PKS.

**_Make sure_** to manually go through the workflow or at least take a look at their [official doc](https://github.com/pivotal-cf/p-pks-integrations/blob/master/CONTRIBUTING.md#fly-your-test-pipelines) firstly before you start. Automation described here only aims to resolve some pain points along the way.

### Step1 Development
Developers make changes in pks-nsx-t-release repo.

### Step2 Create pipeline to build dev release
#### Create profile for release storage options
Firstly, create a `profile` in `control tower` to describe Google Bucket options for your release. For example:
```
➜  pks-nsx-t-release git:(expose-ncpini-in-network-profile) ✗ ct p v -n pks-release-target
pks-nsx-t-release-tarball-bucket: vmw-pks-pipeline-store
pks-nsx-t-release-tarball-path: fangyuanl/pks-nsx-t-(.*).tgz
release_gcs_bucket: vmw-pks-pipeline-store
release_gcs_file: pks-nsx-t-(.*).tgz
release_gcs_file_bash_globbing: pks-nsx-t-*.tgz
release_gcs_path: fangyuanl
release_name: pks-nsx-t
```
Releases built from pipeline will be named and uploaded in above prescribed way.

#### Create pipeline
```
ct d  --template-type build-pks-nsx-t-release --profile-tag kubo -p pks-release-target --profile-path=<(echo -ne "pks-nsx-t-release-branch: expose-ncpini-in-network-profile") --target [target] -n [pipeline name]
```
Above command creates a pipeline `[pipeline name]` in Concourse target `[target]` with `build-pks-nsx-t-release` type. There are three profiles used:
1. profiles with tag `kubo`;
2. profile `pks-release-target` we created before;
3. a profile created on the fly which specifies the git repo branch based on which dev release will be built;

Note: above profile and pipeline can be reused once created. Specify different branch name whenever you'd like to create another dev release.

After pipeline finishes, a dev release will be built, eg: `pks-nsx-t-1.26.0-dev.22`

### Step3 Bump release version
Let's firstly do a quick review of `RAAS`' workstyle describe in doc linked in `Background`.

#### p-pks-integrations
In PKS, [p-pks-integrations](https://github.com/pivotal-cf/p-pks-integrations) repo contains all necessary metadata to build a dev tile. All release versions are captured in a file called [dependencies.yml](https://github.com/pivotal-cf/p-pks-integrations/blob/master/dependencies.yml). To build a dev product with our dev release, we'll need to do following things:
1. create a dev branch on `p-pks-integrations`;
2. update `pks-nsx-t`(or other releases) release metadata in `dependencies.yml` which includes: release version, binary sha etc;

#### pks-releng-ci
 we need to populate file [raas-variables.yml](https://github.com/pivotal-cf/pks-releng-ci/blob/master/raas-variables.yml) with proper values for following variables:
1. feature_name;
2. bucket_name;
3. pks_releng_ci_branch;
4. p_pks_integrations_branch;
5. untested_tile_initial_version

Each component team is given a `team name` by `Release Engineering team`. In our case, it's `nsx-t`.
1. `feature_name` is in the form `<team-name>/<feature-name>` which actually corresponds to a Google Bucket path;
2. `bucket_name` is in the form `<team-name>-test-tile`;
3. `pks_releng_ci_branch` and `p_pks_integrations_branch` are branch names for these two repos;
4. `untested_tile_initial_version` is the base version for your dev tile, if none existed;

It's not hard to see this process is both error-prone and labor-intensive.

Luckily, we've created a pipeline to automate this step:
```
ct d -m bump-release-and-update-branches.yml -p common-secrets -p pks-nsx-t-release-secrets  -p nsx-t-secrets  --profile-tag releng-template -p pks-release-target --target [target] -n [pipeline name]
```
Above command creates a pipeline `[pipeline name]` in Concourse target `[target]` with template [bump-release-and-update-branches.yml](https://gitlab.eng.vmware.com/PKS/pks-concourse/blob/master/pipelines/bump-release-and-update-branches.yml). There are 5 profiles used:
1. common-secrets(static);
2. pks-nsx-t-release-secrets(static);
3. nsx-t-secrets(static);
4. pks-release-target(static);
5. a template profile `releng-template`

First four are static so we'll ignore them here. The last one is a template profile.

For readers who are not familiar with template profile in `control-tower`, if you create a pipeline with a template profile, `control-tower` will interactively ask you values for templated variables. In this case, `ct` will prompt you about values for above 5 variables defined in `raas-variables.yml`. A new profile will be created afterwards with name, tags you specified.

Let's say new profile's name is `master-bump-pks-nsx-t-dev` which sets both branches to be `master`(commonest setting, very likely to be reused in future).

Now, after pipeline finishes, there will be a branch with the same name created both in `p-pks-integrations` and `pks-releng-ci` repo. Eg: `add-pks-nsx-t-1.26.0-dev.22`.

### Step4 Raas Pipelines
From here, you're already fully prepared to fly raas pipelines by:
1. go to `pks-releng-ci` repo locally in terminal;
2. check out above auto-created branch `add-pks-nsx-t-1.26.0-dev.22`;
3. run `./raas-set-pipeline.sh nsx-t [pipeline yaml]` to fly a specific pipeline. eg: `build-tile.yml` will builds a tile for you;

Also, you can choose to create a pipeline to do that.
#### Create profiles for automation credentials
You need two profiles to store `fly` service-account credentials and `raas` credentials for secrets file.
```
➜  pks-nsx-t-release git:(expose-ncpini-in-network-profile) ✗ ct p v -n fly-basic
fly_password: [ask someone]
fly_username: [ask someone]

➜  pipelines git:(master) ct p v -n releng-password
releng-raas-password: [ask someone]
```
#### create pipeline
```
ct d -m fly.yml -p fly-basic -p pks-nsx-t-release-secrets -p releng-password --profile-tag releng-template -p nsx-t-secrets --profile-path <(echo -ne "raas-pipelines: build-tile.yml") --target [target] -n [pipeline name]
```
Above command creates a pipeline `[pipeline name]` in Concourse target `[target]` with template [fly.yml](https://gitlab.eng.vmware.com/PKS/pks-concourse/blob/master/pipelines/fly.yml). 6 profiles are used:
1. we'll ignore static ones here;
2. a profile created on the fly to specify the list of pipelines you want the pipeline to fly on your behalf. Only one key is needed `raas-pipelines`, which takes a string with raas yaml file names separated by `,`. Eg: use "vsphere67-nsx2401-om24-install.yml,vsphere67-nsx2401-om24-upgrade-minor.yml,vsphere67-nsx2401-om24-upgrade-minimum-minor.yml,vsphere67-nsx2401-om25-install.yml" to fly all install and upgrade pipelines.

Note, we use the same template profile again. Branches here should be branches created above automatically to let raas pipelines use updated `dependencies.yml`. As a result, another new profile will be created, eg: `add-pks-nsx-t-1.26.0-dev.22`

### Step5 Install Dev PKS
After tile is built, we'd like to install it in a specified testbed.
```
ct d -p pks-locks-private-key --template-type install-tile -p common-secrets --profile-tag releng  --profile-path=<(echo -ne "install-tile-lock-name: fangyuanl\npks-lock-branch: fangyuanl\npks-lock-pool: nsxt-23-om23") -p add-pks-nsx-t-1.26.0-dev.22
```
Meaning of above command should be clear enough at this point.
Use command:
```
ct f t -j claim-lock
```
to kick it off.

### Summary
1. Every manual step is automated by a pipeline in Concourse;
2. All pipelines created here can be reused;
3. All pipelines created here can be grouped together to work in a fully automated way from beginning to the last installation part;
4. All static profiles created can be reused;
5. Feature specific profiles are created on the fly because of their nature;
