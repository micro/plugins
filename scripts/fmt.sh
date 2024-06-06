#!/bin/bash -ex

version=$1

if [ "x$version" = "x" ]; then
  version='*'
fi

for d in $(find $version -name 'go.mod'); do
  pushd $(dirname $d)
  go mod tidy -go=1.16 && go mod tidy -go=1.17
  go fmt
  popd
done
