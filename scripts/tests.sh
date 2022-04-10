#!/bin/bash -e

for d in $(find * -name 'go.mod'); do
  pushd $(dirname $d) >/dev/null
  go get
  go vet
  #go test -race -v ./... || :
  # go test -v ./...
  popd >/dev/null
done

#go test -race -v $PKGS || :
#go test -v $PKGS
