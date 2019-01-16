package templates

var BuildTileTemplate = `
resource_types:
- name: gcs-resource
  source:
    repository: frodenas/gcs-resource
  type: docker-image

- name: pivnet
  source:
    repository: pivotalcf/pivnet-resource
    tag: latest-final
  type: docker-image

resources:
- name: pks-releng-ci
  source:
    branch: ((pks-releng-ci.branch))
    private_key: ((pks_releng_ci_ssh_key))
    uri: ((pks-releng-ci.uri))
  type: git

- name: download-dependencies-pks-releng-ci
  source:
    branch: ((download-dependencies-pks-releng-ci.branch))
    private_key: ((pks_releng_ci_ssh_key))
    uri: ((download-dependencies-pks-releng-ci.uri))
  type: git

- name: p-pks-integrations
  source:
    branch: ((p-pks-integrations.branch))
    paths:
    - bosh-variables/*
    - forms/*
    - instance-groups/*
    - jobs/*
    - migrations/*
    - properties/*
    - runtime-configs/*
    - base.yml
    - icon.png
    - variables.yml
    - dependencies.yml
    private_key: ((pks_releng_ci_ssh_key))
    uri: ((p-pks-integrations.uri))
  type: git

- name: compilation-lock
  source:
    branch: master
    pool: compilation-locks
    private_key: ((pks-releng-write-locks-bot-key))
    uri: git@github.com:pivotal-cf/pks-releng-ci-locks.git
  type: pool
- name: untested-tile
  source:
    bucket: ((untested-tile.bucket))
    json_key: ((pks-releng-gcp-json))
    regexp: ((untested-tile.regexp))
  type: gcs-resource

- name: product-version
  source:
    bucket: ((untested-tile-version.bucket))
    driver: gcs
    initial_version: ((untested-tile-version.initial_version))
    json_key: ((pks-releng-gcp-json))
    key: ((untested-tile-version.key))
  type: semver

jobs:
- name: build-tile
  plan:
  - aggregate:
    - get: download-dependencies-pks-releng-ci
    - get: pks-releng-ci
    - get: p-pks-integrations
      trigger: true
    - get: product-version
      params:
        pre: build
  - do:
    - params:
        acquire: true
      put: compilation-lock
    - file: download-dependencies-pks-releng-ci/tasks/download-dependencies/task.yml
      input_mapping:
        pks-releng-ci: download-dependencies-pks-releng-ci
      params:
        COMPILED_RELEASES_BUCKET: pks-releng-compiled-releases
        GCP_SERVICE_ACCOUNT_KEY: ((pks-releng-gcp-json))
        PIVNET_TOKEN: ((public-pivnet-token))
      task: download-dependencies
    ensure:
      params:
        release: compilation-lock
      put: compilation-lock
  - config:
      image_resource:
        source:
          repository: pksrelengci/ci-runner
        type: docker-image
      inputs:
      - name: downloads
      platform: linux
      run:
        args:
        - -alhR
        - downloads
        path: ls
    task: list-downloaded-releases
  - file: pks-releng-ci/tasks/validate-tile-inputs/task.yml
    params:
      GCP_SERVICE_ACCOUNT_KEY: ((pks-releng-gcp-json))
    task: check-pks-tile-inputs
  - file: pks-releng-ci/tasks/test-migrations.yml
    task: test-migrations
  - file: pks-releng-ci/tasks/create-tile/task.yml
    task: create-tile
  - do:
    - params:
        file: product-version/number
      put: product-version
    - params:
        content_type: application/octet-stream
        file: build/pivotal-container-service-*.pivotal
      put: untested-tile
  serial: true
`
