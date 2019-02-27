package templates

var RelengAcceptanceTestsPipelineTemplate = `
resource_types:
- name: gcs-resource
  type: docker-image
  source:
    repository: frodenas/gcs-resource
- name: pivnet
  type: docker-image
  source:
    repository: pivotalcf/pivnet-resource
    tag: latest-final
- name: opsmanager
  type: docker-image
  source:
    repository: cflondonservices/opsmanager-resource
    tag: dev
- name: slack-notification
  type: docker-image
  source:
    repository: cfcommunity/slack-notification-resource
    tag: latest
- name: cloudfoundry-bosh-deployment
  type: docker-image
  source:
    repository: cloudfoundry/bosh-deployment-resource
    tag: v2.4.0
- name: pcf-pool
  type: docker-image
  source:
    repository: cftoolsmiths/toolsmiths-envs-resource
    tag: latest
- name: pool-trigger
  type: docker-image
  source:
    repository: cfmobile/pool-trigger
    tag: latest
- name: meta
  type: docker-image
  source:
    repository: swce/metadata-resource

resources:
# Tile related resources
- name: untested-tile
  type: gcs-resource
  source:
    bucket: ((untested-tile.bucket))
    regexp: ((untested-tile.regexp))
    json_key: ((pks-releng-gcp-json))

- name: pks-releng-ci
  type: git
  source:
    uri: ((pks-releng-ci.uri))
    branch: ((pks-releng-ci.branch))
    private_key: ((pks_releng_ci_ssh_key))

- name: p-pks-integrations
  type: git
  source:
    uri: ((p-pks-integrations.uri))
    branch: ((p-pks-integrations.branch))
    private_key: ((pks_releng_ci_ssh_key))

- name: gitlab-pks-nsx-t-release
  type: git
  source:
    uri: git@gitlab.eng.vmware.com:PKS/pks-nsx-t-release.git
    branch: ((gitlab-pks-nsx-t-release.branch))
    private_key: ((gitlab-private-key))

# Lock pool resource
- name: environment-lock
  type: pool
  source:
    branch: ((pks-lock-branch))
    pool: ((pks-lock-pool))
    uri: git@gitlab.eng.vmware.com:PKS/pks-locks.git
    private_key: ((gitlab-private-key))

# Miscellaneous resources
- name: failure-logs
  type: gcs-resource
  source:
    bucket: ((failure-logs.bucket))
    json_key: ((pks-releng-gcp-json))
    regexp: ((failure-logs.regexp))

- name: pipeline-metadata
  type: meta

- name: notify
  type: slack-notification
  source:
    url: ((slack-webhook))

- name: cluster-info-multi-nsx-acceptance-tests
  type: gcs-resource
  source:
    json_key: ((pks-releng-gcp-json))
    bucket: ((cluster-info.bucket))
    regexp: ((cluster-info.regexp))

jobs:
- name: claim-lock
  serial: true
  plan:
  - put: environment-lock
    params:
      claim: ((releng-tests-lock-name))
    timeout: 6h

- name: create-cluster-multi-nsx-acceptance-tests
  serial: true
  plan:
  - aggregate:
    - get: environment-lock
      passed:
      - claim-lock
      trigger: true
    - get: pks-releng-ci
    - get: p-pks-integrations
    - get: untested-tile
  - do:
    - task: download-pks-cli
      file: pks-releng-ci/tasks/download-pks-cli/task.yml
      params:
        GCP_SERVICE_ACCOUNT_KEY: ((pks-releng-gcp-json))
    - task: download-kubectl
      file: pks-releng-ci/tasks/download-kubectl/task.yml
    - task: create-cluster-info-multi-test-tile-deployment
      privileged: true
      file: pks-releng-ci/tasks/create-cluster/task.yml
      params:
        PLAN_NAME: multi-master
        WORKER_INSTANCES: 2
        NETWORK_PROFILE_NAME: pks-network-profile
      output_mapping:
        cluster-info: cluster-info-multi-nsx-acceptance-tests
    - put: cluster-info-multi-nsx-acceptance-tests
      params:
        file: cluster-info-multi-nsx-acceptance-tests/cluster-info-*.yml
    timeout: 2h
  
  on_failure:
    do:
  
  
    - put: notify
      params:
        channel: ((failure-slack-channel))
        attachments:
        - color: danger
          text: $BUILD_PIPELINE_NAME build failed. See results at <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>
  
  
    - get: pipeline-metadata
    - task: download-pks-cli
      file: pks-releng-ci/tasks/download-pks-cli/task.yml
      params:
        GCP_SERVICE_ACCOUNT_KEY: ((pks-releng-gcp-json))
    - task: download-kubectl
      file: pks-releng-ci/tasks/download-kubectl/task.yml
    - task: get-product-version-from-tile
      file: pks-releng-ci/tasks/get-product-version-from-tile.yml
      input_mapping:
        tile: untested-tile
    - task: gather-logs
      privileged: true
      file: pks-releng-ci/tasks/gather-logs.yml
      params:
        ENV_LOCK_FILE: environment-lock/metadata
        IAAS_TYPE: vsphere
    - put: failure-logs
      params:
        file: logs/*-logs.tar.gz
        content_type: application/octet-stream
  

- name: nsx-acceptance-tests
  serial: true
  plan:
  - aggregate:
    - get: environment-lock
      passed:
      - create-cluster-multi-nsx-acceptance-tests
      trigger: true
    - get: untested-tile
      passed:
      - create-cluster-multi-nsx-acceptance-tests
    - get: p-pks-integrations
      passed:
      - create-cluster-multi-nsx-acceptance-tests
    - get: pks-releng-ci
      passed:
      - create-cluster-multi-nsx-acceptance-tests
    - get: gitlab-pks-nsx-t-release
    - get: cluster-info-multi-nsx-acceptance-tests
      passed:
      - create-cluster-multi-nsx-acceptance-tests
      trigger: true
  - do:
    - task: download-pks-cli
      file: pks-releng-ci/tasks/download-pks-cli/task.yml
      params:
        GCP_SERVICE_ACCOUNT_KEY: ((pks-releng-gcp-json))
    - task: download-kubectl
      file: pks-releng-ci/tasks/download-kubectl/task.yml
    - task: nsx-acceptance-tests
      privileged: true
      file: gitlab-pks-nsx-t-release/ci/tasks/run-pks-nsx-t-release-tests.yml
      input_mapping:
        pks-lock: environment-lock
        pks-release: environment-lock
        kubo-deployment: environment-lock
        pks-cluster-info: cluster-info-multi-nsx-acceptance-tests
        git-pks-nsx-t-release: gitlab-pks-nsx-t-release
      params:
        DEPLOYMENT_NAME: test
        TEST_ENVIRONMENT: PKS
        PKS_CLI_USERNAME: alana
        PKS_CLI_PASSWORD: password
        NETWORK_AUTOMATION: true
    timeout: 2h
  
  on_failure:
    do:
  
  
    - put: notify
      params:
        channel: ((failure-slack-channel))
        attachments:
        - color: danger
          text: $BUILD_PIPELINE_NAME build failed. See results at <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>
  
  
    - get: pipeline-metadata
    - task: download-pks-cli
      file: pks-releng-ci/tasks/download-pks-cli/task.yml
      params:
        GCP_SERVICE_ACCOUNT_KEY: ((pks-releng-gcp-json))
    - task: download-kubectl
      file: pks-releng-ci/tasks/download-kubectl/task.yml
    - task: get-product-version-from-tile
      file: pks-releng-ci/tasks/get-product-version-from-tile.yml
      input_mapping:
        tile: untested-tile
    - task: gather-logs
      privileged: true
      file: pks-releng-ci/tasks/gather-logs.yml
      params:
        ENV_LOCK_FILE: environment-lock/metadata
        IAAS_TYPE: vsphere
    - put: failure-logs
      params:
        file: logs/*-logs.tar.gz
        content_type: application/octet-stream
  

- name: release-lock
  plan:
  - get: environment-lock
    passed: [ 'nsx-acceptance-tests' ]
  - put: environment-lock
    params:
      release: environment-lock
`
