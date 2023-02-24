export VERSION := $(shell gogitver)
export KEY_ID := F238881FA1E6EA2A
export ORGANIZATION := iherbllc

current_dir = $(shell pwd)

NAME := paralus
BINARY := terraform-provider-${NAME}
HOSTNAME := hashicorp.com
NAMESPACE := iherbllc

.PHONY: build test testacc install ship terraform-test terraform-apply

build: version
	go build ./...

test: build

testacc:
	# Config JSON should be the PCTL file that is downloaded from the paralus UI. Place it in the same directory as the make file
	PCTL_CONFIG_JSON=$(current_dir)/paralus.local.json CONFIG_JSON=$(current_dir)/paralus.local.json TF_ACC=1 TF_LOG=ERROR go test -v ./internal/acctest

generate_docs:
	go generate ./...

version:
	echo ${VERSION}