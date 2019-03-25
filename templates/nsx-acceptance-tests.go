package templates

var NsxAcceptanceTestsTemplate = `
---
notify_failure: &notify_failure
  on_failure:
    put: notify
    params:
      channel: pks-ci-bots
      attachments:
      - color: danger
        text: $BUILD_PIPELINE_NAME build failed. See results at <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>

release_pks_lock: &release_pks_lock
  put: pks-lock
  params:
    release: pks-lock

delete_kubo: &delete_kubo
  task: delete-kubo
  file: git-pks-ci/ci/tasks/delete-kubo.yml
  input_mapping:
    git-pks-ci: git-pks-ci
    pks-lock: pks-lock
    pks-release: gcs-pks-nsx-t-tarball-untested
    kubo-deployment: kubeconfig
  params:
    DEPLOYMENT_NAME: kubo-pks-nsx-t
    CLEANUP_NSX: true
    USE_NAT: false

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
- name: p-pks-integrations
  type: git
  source:
    uri: ((p-pks-integrations.uri))
    branch: master
    private_key: ((pks_releng_ci_ssh_key))

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

# pks-releng-ci git repo
- name: pks-releng-ci
  type: git
  source:
    uri: ((pks-releng-ci.uri))
    branch: ((pks-releng-ci.branch))
    private_key: ((pks_releng_ci_ssh_key))

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

# test bed lock resource
- name: pks-lock
  type: pool
  source:
    uri: git@locks.pks.eng.vmware.com:pks-locks.git
    branch: ((pks-lock-branch))
    private_key: ((gitlab-private-key))
    pool: ((pks-lock-pool))

# temp bucket to upload the release tars to
- name: gcs-pks-nsx-t-tarball-untested
  type: gcs
  source:
    json_key: ((gcs-json-key))
    bucket: vmw-pks-pipeline-store
    regexp: pks-nsx-t/pks-nsx-t-(.*).tgz

- name: gcs-nsx-cf-cni-tarball
  type: gcs
  source:
    json_key: ((gcs-json-key))
    bucket: vmw-pks-releases
    regexp: nsx-cf-cni/2.3.1/nsx-cf-cni-(.*).tgz

# stemcell
- name: ubuntu-xenial-stemcell
  type: bosh-io-stemcell
  source:
    name: bosh-vsphere-esxi-ubuntu-xenial-go_agent

# kubo-deployment github repo
- name: github-kubo-deployment
  type: github-release
  source:
    owner: cloudfoundry-incubator
    repository: kubo-deployment
    access_token: ((github-access-token))
    release: true
    pre_release: true

- name: kubeconfig
  type: gcs
  source:
    json_key: ((gcs-json-key))
    bucket: ((kubeconfig-bucket))
    regexp: ((kubeconfig-folder))

############################################
# Groups
############################################
groups:
  - name: all
    jobs:
      - claim-lock
      - run-release-tests
      - claim-lock-conformance-tests
      - run-conformance-tests

  - name: release-tests
    jobs:
      - claim-lock
      - run-release-tests
      - run-release-tests-release-lock
      - run-release-tests-delete-kubo

  - name: conformance-tests
    jobs:
      - claim-lock-conformance-tests
      - run-conformance-tests
      - run-conformance-tests-release-lock
      - run-conformance-tests-delete-kubo

############################################
# Jobs
############################################
jobs:

# Claim kubo test bed
- name: claim-lock
  serial: true
  plan:
  - put: pks-lock
    params:
      claim: ((lock-name))

  <<: *notify_failure

- name: run-release-tests-release-lock
  plan:
  - get: pks-lock
    passed: [ 'claim-lock' ]
  - <<: *release_pks_lock

- name: run-release-tests-delete-kubo
  serial: true
  serial_groups:
    - kubo-pks-nsx-t
  plan:
  - aggregate:
    - get: git-pks-ci
    - get: gcs-pks-nsx-t-tarball-untested
    - get: gcs-nsx-cf-cni-tarball
    - get: pks-lock
      passed: [ 'claim-lock' ]
      version: every
    - get: ubuntu-xenial-stemcell
    - get: kubeconfig

  - <<: *delete_kubo

  <<: *notify_failure

# Run release tests
- name: run-release-tests
  serial: true
  serial_groups:
    - kubo-pks-nsx-t
  plan:
  - aggregate:
    - get: pks-releng-ci
    - get: p-pks-integrations
    - get: git-pks-nsx-t-release
      trigger: true
    - get: pks-nsx-t-version
      trigger: true
    - get: gcs-pks-nsx-t-tarball-untested
    - get: gcs-nsx-cf-cni-tarball
      trigger: true
    - get: pks-lock
      version: every
      passed: [ 'claim-lock' ]
      trigger: true
    - get: ubuntu-xenial-stemcell
    - get: github-kubo-deployment
    - get: kubeconfig

  - task: download-kubectl
    file: pks-releng-ci/tasks/download-kubectl/task.yml

  - task: run-tests
    file: git-pks-nsx-t-release/ci/tasks/run-pks-nsx-t-release-tests.yml
    input_mapping:
      pks-lock: pks-lock
      pks-release: gcs-pks-nsx-t-tarball-untested
      ncp-release: gcs-nsx-cf-cni-tarball
      pks-cli: pks-lock
      kubo-deployment: kubeconfig
      git-pks-nsx-t-release: git-pks-nsx-t-release
    params:
      RELEASE_NAME: pks-nsx-t
      NCP_RELEASE_NAME: nsx-cf-cni
      DEPLOYMENT_NAME: kubo-pks-nsx-t
      TEST_ENVIRONMENT: kubo
      NETWORK_AUTOMATION: false
      MULTI_MASTER: false

- name: claim-lock-conformance-tests
  serial: true
  plan:
  - put: pks-lock
    params:
      claim: ((lock-name))

  <<: *notify_failure

- name: run-conformance-tests-release-lock
  plan:
  - get: pks-lock
    passed: [ 'claim-lock-conformance-tests' ]
  - <<: *release_pks_lock

# Run release tests
- name: run-conformance-tests-delete-kubo
  serial: true
  serial_groups:
    - kubo-pks-nsx-t
  plan:
  - aggregate:
    - get: git-pks-ci
    - get: gcs-pks-nsx-t-tarball-untested
    - get: gcs-nsx-cf-cni-tarball
    - get: pks-lock
      passed: [ 'claim-lock-conformance-tests' ]
      version: every
    - get: ubuntu-xenial-stemcell
    - get: kubeconfig

  - <<: *delete_kubo
  <<: *notify_failure

# Run release tests
- name: run-conformance-tests
  serial: true
  serial_groups:
    - kubo-pks-nsx-t
  plan:
  - aggregate:
    - get: git-pks-ci
      trigger: true
    - get: git-pks-nsx-t-release
      trigger: true
    - get: pks-nsx-t-version
      trigger: true
    - get: gcs-pks-nsx-t-tarball-untested
    - get: gcs-nsx-cf-cni-tarball
      trigger: true
    - get: pks-lock
      passed: [ 'claim-lock-conformance-tests' ]
      version: every
      trigger: true
    - get: ubuntu-xenial-stemcell
    - get: github-kubo-deployment
    - get: kubeconfig
    - get: pks-releng-ci

  - task: run-k8s-conformance-tests
    file: git-pks-ci/ci/tasks/run-k8s-conformance-tests.yml
    input_mapping:
      pks-releng-ci: pks-releng-ci
      version: pks-nsx-t-version
      git-pks-ci: git-pks-ci
      lock: pks-lock
      kubeconfig: kubeconfig
`
