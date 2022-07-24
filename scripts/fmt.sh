#!/bin/bash -ex

for d in $(find * -name 'go.mod'); do
  pushd $(dirname $d)
  go fmt
  # go mod tidy
  popd
done
