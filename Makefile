
build:
	go build -o bin/task-runner main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/task-runner-linux main.go

deploy:
	scp ./bin/task-runner-linux root@192.168.1.225:/usr/bin/task-runner
