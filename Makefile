build: build-ct move-to-path

build-ct:
	go build -o ct main.go

move-to-path:
	mv ct /usr/local/bin
