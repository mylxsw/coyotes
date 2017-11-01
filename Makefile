run:build-mac
	./bin/coyotes -colorful-tty=true -debug=true
run-with-backend:build-mac
	./bin/coyotes -colorful-tty=true -debug=true -backend-storage="mysql:root:@tcp(127.0.0.1:3306)/coyotes?charset=utf8&parseTime=True&loc=Local" -backend-keep-days=1
run-no-worker:build-mac
	./bin/coyotes -colorful-tty=true -task-mode=false -debug=true
run-redis-230:build-mac
	./bin/coyotes -colorful-tty=true -debug=true -host 192.168.1.230:6379

run-race-check:
	go run -race *.go -colorful-tty=true -debug=true -backend-storage="mysql:root:@tcp(127.0.0.1:3306)/coyotes?charset=utf8&parseTime=True&loc=Local" -backend-keep-days=1

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
