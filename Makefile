
build:
	go build 

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

deploy:
	scp ./task-runner root@192.168.1.225:/usr/bin/task-runner
