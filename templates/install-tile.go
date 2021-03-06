package templates

var InstallTileTemplate = `
---
release_lock: &release_lock
  put: environment-lock
  params:
    release: environment-lock

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

- name: kubo-odb-ci
  type: git
  source:
    branch: ((kubo-odb-ci.branch))
    private_key: ((pks_releng_ci_ssh_key))
    uri: ((kubo-odb-ci.uri))

- name: p-pks-integrations
  type: git
  source:
    uri: ((p-pks-integrations.uri))
    branch: ((p-pks-integrations.branch))
    private_key: ((pks_releng_ci_ssh_key))

- name: git-tile-pipeline
  type: git
  source:
    branch: ((git-tile-pipeline.branch))
    private_key: ((pks_releng_ci_ssh_key))
    tag_filter: ((git-tile-pipeline.tag_filter))
    uri: ((git-tile-pipeline.uri))

# Lock pool resource
- name: environment-lock
  type: pool
  source:
    branch: ((pks-lock-branch))
    pool: ((pks-lock-pool))
    uri: git@locks.pks.eng.vmware.com:pks-locks.git
    private_key: ((pks-locks-private-key))

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

- name: pivnet-stemcell
  type: pivnet
  source:
    api_token: ((public-pivnet-token))
    product_slug: stemcells-ubuntu-xenial
    product_version: (\d)*\.(\d)*
    sort_by: semver

jobs:
- name: claim-lock
  serial: true
  plan:
  - put: environment-lock
    params:
      claim: ((install-tile-lock-name))
    timeout: 6h

- name: ensure-clean-environment
  serial: true
  plan:
  - aggregate:
    - get: untested-tile
    - get: p-pks-integrations
    - get: pks-releng-ci
    - get: environment-lock
      passed:
      - claim-lock
      trigger: true
  - task: ensure-clean-environment
    privileged: true
    file: pks-releng-ci/tasks/ensure-clean-environment/task.yml
  on_failure:
    do:
    - put: notify
      params:
        channel: ((failure-slack-channel))
        attachments:
        - color: danger
          text: $BUILD_PIPELINE_NAME build failed. See results at <https://((ci_url))/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME>

- name: ensure-clean-environment-release-lock
  serial: true
  plan:
  - get: environment-lock
    passed: [ 'claim-lock' ]
  - <<: *release_lock

- name: add-tile
  serial: true
  plan:
  - aggregate:
    - get: environment-lock
      trigger: true
      passed:
      - ensure-clean-environment
    - get: untested-tile
      passed:
      - ensure-clean-environment
    - get: p-pks-integrations
      passed:
      - ensure-clean-environment
    - get: pks-releng-ci
      passed:
      - ensure-clean-environment
  - do:
    - task: add-tile-to-opsman
      privileged: true
      file: pks-releng-ci/tasks/add-tile-to-opsman.yml
      params:
        ENV_LOCK_FILE: environment-lock/metadata
        PIVNET_TOKEN: ((public-pivnet-token))
    timeout: 1h
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

- name: add-tile-release-lock
  serial: true
  plan:
  - get: environment-lock
    passed: [ 'ensure-clean-environment' ]
  - <<: *release_lock

- name: configure-and-deploy-tile
  serial: true
  plan:
  - aggregate:
    - get: environment-lock
      passed:
      - add-tile
      trigger: true
    - get: untested-tile
      passed:
      - add-tile
    - get: kubo-odb-ci
    - get: p-pks-integrations
      passed:
      - add-tile
    - get: pks-releng-ci
      passed:
      - add-tile
    - get: git-tile-pipeline
    - get: pivnet-stemcell
  - task: get-product-version-from-tile
    file: pks-releng-ci/tasks/get-product-version-from-tile.yml
    input_mapping:
      tile: untested-tile
  - task: create-tile-configuration
    file: pks-releng-ci/tasks/create-tile-configuration/task.yml
    privileged: true
    params:
      NETWORK_AUTOMATION: true
      VRLI_ENABLED: true
  - task: deploy-mountebank
    file: pks-releng-ci/tasks/deploy-mountebank/task.yml
    privileged: true
    params:
      PIVNET_TOKEN: ((public-pivnet-token))
  - task: extract-complex-secrets
    file: pks-releng-ci/tasks/extract-complex-secrets/task.yml
  - task: generate-om-certificate
    file: pks-releng-ci/tasks/generate-om-certificate/task.yml
    privileged: true
  - task: configure-product
    file: pks-releng-ci/tasks/configure-product/task.yml
    privileged: true
  - task: apply-changes
    privileged: true
    file: pks-releng-ci/tasks/apply-changes/task.yml
  - task: configure-uaa
    privileged: true
    file: pks-releng-ci/tasks/setup-uaa/task.yml
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

- name: configure-and-deploy-tile-release-lock
  serial: true
  plan:
  - get: environment-lock
    passed: [ 'add-tile' ]
  - <<: *release_lock

- name: release-lock
  serial: true
  plan:
  - get: environment-lock
    passed: [ 'configure-and-deploy-tile' ]
    trigger: true
  - <<: *release_lock

`
