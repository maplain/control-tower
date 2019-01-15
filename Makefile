build: build-ct move-to-path

build-ct:
	go build -o ct main.go

move-to-path:
	mv ct /usr/local/bin

set-raas:
	hack/setup_raas setup

set-raas-force:
	hack/setup_raas setup --overwrite
