build: build-ct move-to-path

build-ct:
	go build -o bin/ct main.go

move-to-path:
	mv bin/ct /usr/local/bin

set-raas:
	hack/setup_raas setup

set-raas-force:
	hack/setup_raas setup --overwrite

rm-raas:
	hack/setup_raas teardown

update-kubo-static:
	hack/setup_kubo setup --overwrite --skip

set-kubo:
	hack/setup_kubo setup

set-kubo-force:
	hack/setup_kubo setup --overwrite

rm-kubo:
	hack/setup_kubo teardown
