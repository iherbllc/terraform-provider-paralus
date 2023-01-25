export VERSION := $(shell gogitver)
export ORGANIZATION := iherbllc

current_dir = $(shell pwd)

NAME := paralus
BINARY := terraform-provider-${NAME}
HOSTNAME := hashicorp.com
NAMESPACE := iherbllc

.PHONY: build test testacc install ship terraform-test terraform-apply tag push

build: version
	go build ./...

test: build

testacc:
	# Config JSON should be the PCTL file that is downloaded from the paralus UI. Place it in the same directory as the make file
	CONFIG_JSON=$(current_dir)/config.json TF_ACC=1 TF_LOG=ERROR go test -v ./internal/acctest

version:
	echo ${VERSION}

tag:
	git tag v${VERSION}

push:
	git push origin v${VERSION}