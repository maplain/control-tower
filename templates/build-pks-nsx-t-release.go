package templates

// ct d  --template-type build-pks-nsx-t-release --profile-tag kubo -p pks-release-target --profile-path=<(echo -ne "pks-nsx-t-release-branch: expose-ncpini-in-network-profile")

var BuildPksNSXTReleaseTemplate = `
---
resource_types:
- name: gcs
  type: docker-image
  source:
    repository: frodenas/gcs-resource

- name: slack-notification
  type: docker-image
  source:
    repository: cfcommunity/slack-notification-resource
    tag: latest

resources:
# slack
- name: notify
  type: slack-notification
  source:
    url: ((slack-webhook))

# pks ci git repo
- name: git-pks-ci
  type: git
  source:
    uri: git@github.com:vmware/pks-ci.git
    branch: master
    private_key: ((github-private-key))

# pks-nsx-t-release git repo
- name: git-pks-nsx-t-release
  type: git
  source:
    uri: git@gitlab.eng.vmware.com:PKS/pks-nsx-t-release.git
    branch: ((pks-nsx-t-release-branch))
    private_key: ((gitlab-private-key))

# pks-nsx-t-release version
- name: pks-nsx-t-version
  type: semver
  source:
    driver: gcs
    key: pks-nsx-t-version
    json_key: ((gcs-json-key))
    bucket: vmw-pks-pipeline-store

# temp bucket to upload the release tars to
- name: gcs-pks-nsx-t-tarball-untested
  type: gcs
  source:
    json_key: ((gcs-json-key))
    bucket: ((pks-nsx-t-release-tarball-bucket))
    regexp: ((pks-nsx-t-release-tarball-path))

############################################
# Groups
############################################
groups:
  - name: all
    jobs:
      - run-security-checks
      - run-unit-tests
      - build-dev-release

############################################
# Jobs
############################################
jobs:

# Run security checks
- name: run-security-checks
  serial: true
  plan:
    - aggregate:
        - get: git-pks-nsx-t-release
          trigger: true
    - task: run-security-checks
      file: git-pks-nsx-t-release/ci/tasks/run-pks-nsx-t-security-checks.yml
      input_mapping:
        git-pks-release: git-pks-nsx-t-release
  on_failure:
    put: notify
    params:
      channel: pks-ci-bots
      attachments:
      - color: danger
        text: $BUILD_PIPELINE_NAME build failed. See results at <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>


# Run unit tests
- name: run-unit-tests
  serial: true
  plan:
  - aggregate:
    - get: git-pks-ci
    - get: git-pks-nsx-t-release
      passed: [ 'run-security-checks' ]
      trigger: true
  - task: run-osb-proxy-unit-tests
    file: git-pks-nsx-t-release/ci/tasks/run-pks-nsx-t-osb-proxy-unit-tests.yml
    input_mapping:
      git-pks-release: git-pks-nsx-t-release
  - task: run-pks-nsx-t-release-unit-tests
    file: git-pks-nsx-t-release/ci/tasks/run-pks-nsx-t-release-unit-tests.yml
    input_mapping:
      git-pks-release: git-pks-nsx-t-release
  on_failure:
    put: notify
    params:
      channel: pks-ci-bots
      attachments:
      - color: danger
        text: $BUILD_PIPELINE_NAME build failed. See results at <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>


# Build dev release
- name: build-dev-release
  serial: true
  plan:
  - aggregate:
    - get: git-pks-ci
      passed: [ 'run-unit-tests' ]
      trigger: true
    - get: git-pks-nsx-t-release
      passed: [ 'run-unit-tests' ]
      trigger: true
    - get: pks-nsx-t-version
      params:
        pre: dev
  - task: build-dev-release
    file: git-pks-ci/ci/tasks/build-dev-release.yml
    input_mapping:
      git-pks-release: git-pks-nsx-t-release
      pks-version: pks-nsx-t-version
    params:
      GCS_ACCESS_KEY:  ((gcs-access-key))
      GCS_SECRET_KEY:  ((gcs-secret-key))
  - put: gcs-pks-nsx-t-tarball-untested
    params:
      file: pks-release/pks-nsx-t-*.tgz
  - put: pks-nsx-t-version
    params:
      file: pks-nsx-t-version/version
  on_failure:
    put: notify
    params:
      channel: pks-ci-bots
      attachments:
      - color: danger
        text: $BUILD_PIPELINE_NAME build failed. See results at <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>
`
