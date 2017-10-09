#!/bin/bash

set -e

if [[ $TRAVIS_OS_NAME == "osx" ]]; then
  wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-osx-amd64
  sudo mv jq-osx-amd64 /usr/local/bin/jq
  jq --version
fi

# Lets install the dependencies that are not vendored in anymore.
go get -d golang.org/x/net/context
go get -d google.golang.org/grpc/...


