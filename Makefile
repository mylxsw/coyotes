run:build-mac
	./bin/task-runner -colorful-tty=true
run-no-worker:build-mac
	./bin/task-runner -colorful-tty=true -task-mode=false
run-redis-230:build-mac
	./bin/task-runner -colorful-tty=true -host 192.168.1.230:6379

build-mac:
	go build -o bin/task-runner *.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/task-runner-linux *.go

deploy-mac:build-mac
	cp ./bin/task-runner /usr/local/bin/task-runner

clean-linux:
	rm -fr ./bin/task-runner-linux

clean-mac:
	rm -fr ./bin/task-runner

clean:clean-linux clean-mac

include Makefile.local
