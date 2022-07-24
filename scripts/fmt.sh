#!/bin/bash -ex

version=$1

if [ "x$version" = "x" ]; then
  version='*'
fi

for d in $(find $version -name 'go.mod'); do
  pushd $(dirname $d)
  go fmt
  go mod tidy
  popd
done
