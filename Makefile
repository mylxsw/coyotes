run-server:build-mac
	./bin/coyotes-server -colorful-tty=true -debug=true
run-node:build-mac
	./bin/coyotes-node 
run-redis-230:build-mac
	./bin/coyotes-server -colorful-tty=true -debug=true -host 192.168.1.230:6379

build-mac:
	go build -o bin/coyotes-server coyotes/server/*.go
	go build -o bin/coyotes-node coyotes/node/*.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/coyotes-server-linux coyotes/server/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/coyotes-node-linux coyotes/node/*.go

deploy-mac:build-mac
	cp ./bin/coyotes-server /usr/local/bin/coyotes-server
	cp ./bin/coyotes-node /usr/local/bin/coyotes-node

clean-linux:
	rm -fr ./bin/coyotes-server-linux
	rm -fr ./bin/coyotes-node-linux

clean-mac:
	rm -fr ./bin/coyotes-server
	rm -fr ./bin/coyotes-node

clean:clean-linux clean-mac

include Makefile.local
