export GOPATH := $(shell pwd)
export PATH := $(GOPATH)/bin:$(PATH)

build:
	@echo "--> go get..."
	go get github.com/XeLabs/go-mysqlstack/driver

	@echo "--> Building..."
	@mkdir -p bin/
	go build -v -o bin/benchyou src/bench/benchyou.go
	@chmod 755 bin/*

clean:
	@echo "--> Cleaning..."
	@go clean
	@rm -f bin/*

fmt:
	go fmt ./...
