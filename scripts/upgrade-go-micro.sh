#!/bin/bash -ex

version=$1
gomicro_version=$2

if [ -z "$version" ]; then
    version='*'
fi

if [ -z "$gomicro_version" ]; then
    gomicro_version='latest'
fi

for d in $(find $version -name 'go.mod'); do
    pushd $(dirname $d)
    go get go-micro.dev/v4@"$gomicro_version"
    go mod tidy
    popd
done
