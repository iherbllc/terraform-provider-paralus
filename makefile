export VERSION := $(shell gogitver)
export KEY_ID := F238881FA1E6EA2A
export ORGANIZATION := iherbllc

NAME := bitbucket-file
BINARY := terraform-provider-${NAME}
HOSTNAME := hashicorp.com
NAMESPACE := iherbllc
NAME := bitbucket-file

.PHONY: build test install ship terraform-test terraform-apply

build: version
	go build ./...

test: build

version:
	echo ${VERSION}
