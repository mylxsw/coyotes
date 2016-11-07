
build-mac:
	go build -o bin/task-runner main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/task-runner-linux main.go

deploy-mac:build-mac
	cp ./bin/task-runner /usr/local/bin/task-runner

clean-linux:
	rm -fr ./bin/task-runner-linux

clean-mac:
	rm -fr ./bin/task-runner

clean:clean-linux clean-mac

include Makefile.local
