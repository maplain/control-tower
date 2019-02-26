package templates

var DeployKuboPipelineTemplate = `
---
notify_failure: &notify_failure
  on_failure:
    put: notify
    params:
      channel: pks-ci-bots
      attachments:
      - color: danger
        text: $BUILD_PIPELINE_NAME build failed. See results at <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>

notify_success: &notify_success
  on_success:
    put: notify
    params:
      channel:  pks-ci-bots
      attachments:
      - color: good
        text: $BUILD_PIPELINE_NAME build succeeded <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>

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
    kubo-deployment: kubo-deployment
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

# pks-networking git repo
- name: git-pks-networking
  type: git
  source:
    uri: git@gitlab.eng.vmware.com:PKS/pks-networking.git
    branch: master
    private_key: ((gitlab-private-key))

# pks-nsx-t-release git repo
- name: git-pks-nsx-t-release
  type: git
  source:
    uri: git@gitlab.eng.vmware.com:PKS/pks-nsx-t-release.git
    branch: ((pks-nsx-t-release-branch))
    private_key: ((gitlab-private-key))

# pks-concourse git repo
- name: pks-concourse
  type: git
  source:
    uri: git@gitlab.eng.vmware.com:PKS/pks-concourse.git
    branch: master
    private_key: ((pks-concourse-private-key))

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
    uri: git@gitlab.eng.vmware.com:PKS/pks-locks.git
    branch: ((pks-lock-branch))
    private_key: ((gitlab-private-key))
    pool: ((pks-lock-pool))

# temp bucket to upload the release tars to
- name: gcs-pks-nsx-t-tarball-untested
  type: gcs
  source:
    json_key: ((gcs-json-key))
    bucket: ((pks-nsx-t-release-tarball-bucket))
    regexp: ((pks-nsx-t-release-tarball-path))

- name: gcs-nsx-cf-cni-tarball
  type: gcs
  source:
    json_key: ((gcs-json-key))
    bucket: vmw-pks-releases
    regexp: nsx-cf-cni/2.3.2/nsx-cf-cni-(.*).tgz

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

# kubo-release github repo
- name: github-kubo-release
  type: github-release
  source:
    owner: cloudfoundry-incubator
    repository: kubo-release
    access_token: ((github-access-token))
    release: true
    pre_release: true

# kubeconfig version
- name: kubeconfig-version
  type: semver
  source:
    driver: gcs
    key: kubeconfig-version
    json_key: ((gcs-json-key))
    bucket: vmw-pks-pipeline-store

- name: kubeconfig
  type: gcs
  source:
    json_key: ((gcs-json-key))
    bucket: vmw-pks-pipeline-store
    regexp: fangyuanl/kubeconfig-(.*).tgz

############################################
# Groups
############################################
groups:
  - name: all
    jobs:
      - claim-lock-kubo
      - run-precheck
      - run-precheck-release-lock
      - deploy-kubo
      - deploy-kubo-release-lock

  - name: outputs
    jobs:
      - outputs
      - claim-lock-for-outputs
      - outputs-release-lock

  - name: delete-kubo
    jobs:
      - claim-lock-for-deletion
      - delete-kubo
      - delete-kubo-release-lock

  - name: dev
    jobs:
      - claim-lock-kubo
      - run-precheck
      - deploy-kubo
      - delete-kubo
      - outputs

############################################
# Jobs
############################################
jobs:

# Claim kubo test bed
- name: claim-lock-kubo
  serial: true
  plan:
  - put: pks-lock
    params:
      acquire: true
  <<: *notify_failure


- name: run-precheck-release-lock
  serial_groups:
    - run-precheck
  serial: true
  plan:
  - get: pks-lock
    passed: [ 'claim-lock-kubo' ]
  - <<: *release_pks_lock

# Run precheck
- name: run-precheck
  serial: true
  serial_groups:
    - run-precheck
  plan:
  - aggregate:
    - get: git-pks-ci
      trigger: true
    - get: git-pks-networking
    - get: git-pks-nsx-t-release
      trigger: true
    - get: pks-nsx-t-version
      trigger: true
    - get: gcs-pks-nsx-t-tarball-untested
    - get: pks-lock
      trigger: true
      passed: [ 'claim-lock-kubo' ]
      trigger: true
      version: every
    - get: ubuntu-xenial-stemcell

  - task: run-precheck
    file: git-pks-nsx-t-release/ci/tasks/run-pks-nsx-t-precheck.yml
    input_mapping:
      git-pks-ci: git-pks-ci
      git-pks-networking: git-pks-networking
      pks-lock: pks-lock
      git-pks-release: git-pks-nsx-t-release
      pks-release: gcs-pks-nsx-t-tarball-untested
      stemcell: ubuntu-xenial-stemcell
    params:
      RELEASE_NAME: pks-nsx-t
      DEPLOYMENT_NAME: pks-nsx-t

  <<: *notify_failure


- name: deploy-kubo-release-lock
  plan:
  - get: pks-lock
    passed: [ 'deploy-kubo' ]
  - <<: *release_pks_lock

# Claim kubo test bed
- name: claim-lock-for-outputs
  serial: true
  serial_groups:
    - outputs
  plan:
  - put: pks-lock
    params:
      acquire: true
  <<: *notify_failure

- name: outputs
  serial: true
  serial_groups:
    - outputs
  plan:
  - aggregate:
    - get: pks-lock
      passed: [ 'claim-lock-for-outputs' ]
      trigger: true
    - get: kubeconfig
      passed: [ 'deploy-kubo' ]
      trigger: true
    - get: pks-concourse
      trigger: true
  - task: outputs
    file: pks-concourse/tasks/deploy-kubo-outputs.yml

- name: outputs-release-lock
  serial: true
  serial_groups:
    - outputs
  plan:
  - get: pks-lock
    passed: [ 'claim-lock-for-outputs' ]
  - <<: *release_pks_lock

- name: deploy-kubo
  serial: true
  serial_groups:
    - deploy-kubo
  plan:
  - aggregate:
    - get: git-pks-ci
      passed: [ 'run-precheck' ]
      trigger: true
    - get: git-pks-nsx-t-release
      passed: [ 'run-precheck' ]
      trigger: true
    - get: pks-nsx-t-version
      passed: [ 'run-precheck' ]
      trigger: true
    - get: gcs-pks-nsx-t-tarball-untested
      passed: [ 'run-precheck' ]
    - get: gcs-nsx-cf-cni-tarball
      trigger: true
    - get: pks-lock
      passed: [ 'run-precheck' ]
      version: every
      trigger: true
    - get: ubuntu-xenial-stemcell
      passed: [ 'run-precheck' ]
    - get: github-kubo-deployment
    - get: github-kubo-release
    - put: kubeconfig-version
      params: {bump: minor}

  - task: deploy-kubo
    file: git-pks-ci/ci/tasks/deploy-kubo.yml
    input_mapping:
      git-pks-ci: git-pks-ci
      pks-lock: pks-lock
      github-kubo-deployment: github-kubo-deployment
      github-kubo-release: github-kubo-release
      git-pks-release: git-pks-nsx-t-release
      pks-release: gcs-pks-nsx-t-tarball-untested
      ncp-release: gcs-nsx-cf-cni-tarball
      stemcell: ubuntu-xenial-stemcell
      kubeconfig-version: kubeconfig-version
    params:
      RELEASE_NAME: pks-nsx-t
      NCP_RELEASE_NAME: nsx-cf-cni
      DEPLOYMENT_NAME: kubo-pks-nsx-t
      DEPLOYMENT_OP_FILES: git-pks-release/manifests/operators/add_pks_nsx_t.yml,git-pks-release/manifests/operators/prepare_master_vm.yml,git-pks-release/manifests/operators/add_pks_nsx_t_floating_ip_association.yml,git-pks-release/manifests/operators/remove_flannel.yml
      K8S_WORKER_COUNT: 3
      ENABLE_BOSH_DNS_OVERRIDE_NAMESERVER: true
      CLEAN_UP_FAILED_RUNS: true
      USE_NAT: false
  - put: kubeconfig
    params:
      file: kubo-deployment/kubeconfig-*.tgz

  on_failure:
    <<: *delete_kubo
  on_abort:
    <<: *delete_kubo
  <<: *notify_success

# Claim kubo test bed
- name: claim-lock-for-deletion
  serial: true
  plan:
  - put: pks-lock
    params:
      acquire: true
  <<: *notify_failure

- name: delete-kubo
  serial: true
  serial_groups:
    - delete-kubo
  plan:
  - aggregate:
    - get: git-pks-ci
      passed: [ 'deploy-kubo' ]
    - get: git-pks-nsx-t-release
      passed: [ 'deploy-kubo' ]
    - get: pks-nsx-t-version
      passed: [ 'deploy-kubo' ]
    - get: gcs-pks-nsx-t-tarball-untested
      passed: [ 'deploy-kubo' ]
    - get: gcs-nsx-cf-cni-tarball
    - get: pks-lock
      passed: [ 'claim-lock-for-deletion' ]
      version: every
    - get: kubeconfig
      passed: [ 'deploy-kubo' ]
  - task: delete-kubo
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

    on_failure:
      <<: *release_pks_lock
    on_abort:
      <<: *release_pks_lock
  <<: *notify_failure

- name: delete-kubo-release-lock
  serial: true
  serial_groups:
    - delete-kubo
  plan:
  - get: pks-lock
    passed: [ 'claim-lock-for-deletion' ]
  - <<: *release_pks_lock
`
