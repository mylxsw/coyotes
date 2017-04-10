run:build-mac
	./bin/coyotes -colorful-tty=true
run-no-worker:build-mac
	./bin/coyotes -colorful-tty=true -task-mode=false
run-redis-230:build-mac
	./bin/coyotes -colorful-tty=true -host 192.168.1.230:6379

build-mac:
	go build -o bin/coyotes *.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/coyotes-linux *.go

deploy-mac:build-mac
	cp ./bin/coyotes /usr/local/bin/coyotes

clean-linux:
	rm -fr ./bin/coyotes-linux

clean-mac:
	rm -fr ./bin/coyotes

clean:clean-linux clean-mac

include Makefile.local
