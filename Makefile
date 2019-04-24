#-----------------------------------------------------------------------------
# Global Variables
#-----------------------------------------------------------------------------
GO   := go

DOCKER_USER ?= $(DOCKER_USER)
DOCKER_PASS ?= 

DOCKER_BUILD_ARGS := --build-arg HTTP_PROXY=$(http_proxy) --build-arg HTTPS_PROXY=$(https_proxy)

APP_VERSION := latest

GOLANGCI:=$(shell command -v golangci-lint 2> /dev/null)

pkgs  = $(shell $(GO) list ./... | grep -vE -e /vendor/)
pkgDirs = $(shell $(GO) list -f {{.Dir}} ./... | grep -vE -e /vendor/)

#-----------------------------------------------------------------------------
# BUILD
#-----------------------------------------------------------------------------

.PHONY: default build test publish build_local lint
default: depend test lint build 

depend:
	go get -u github.com/golang/dep
	dep ensure -vendor-only
test:
	go test -v ./...
build_local:
	go build 
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	docker build $(DOCKER_BUILD_ARGS) -t $(DOCKER_USER)/k8sslackevent:$(APP_VERSION)  .

lint:
ifndef GOLANGCI
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
endif
	@golangci-lint run -v $(pkgDirs)

format:
	@echo "==> formatting code"
	@$(GO) fmt $(pkgs)
	@echo "==> clean imports"
	@goimports -w $(pkgDirs)
	@echo "==> simplify code"
	@gofmt -s -w $(pkgDirs)

#-----------------------------------------------------------------------------
# PUBLISH
#-----------------------------------------------------------------------------

.PHONY: publish 

publish: 
	docker push $(DOCKER_USER)/k8sslackevent:$(APP_VERSION)

#-----------------------------------------------------------------------------
# CLEAN
#-----------------------------------------------------------------------------

.PHONY: clean 

clean:
	rm -rf k8sslackevent