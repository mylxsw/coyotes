
build-mac:
	go build -o bin/task-runner main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/task-runner-linux main.go

deploy:build-linux
	scp ./bin/task-runner-linux root@192.168.1.225:/usr/bin/task-runner
	scp ./bin/task-runner-linux root@192.168.1.226:/usr/bin/task-runner

deploy-mac:build-mac
	cp ./bin/task-runner /usr/local/bin/task-runner

clean-linux:
	rm -fr ./bin/task-runner-linux

clean-mac:
	rm -fr ./bin/task-runner

clean:clean-linux clean-mac
