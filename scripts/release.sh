#!/bin/bash

version=$1
tag=$2
commitsh=$3

if [ "x$version" = "x" ]; then
  echo "must specify version to release"
  exit 1;
fi

if [ "x$tag" = "x" ]; then
  echo "must specify tag to release"
  exit 1;
fi

for m in $(find $version -name 'go.mod' -exec dirname {} \;); do
  if [ ! -n "$commitsh" ]; then
    hub release create -m "$m/$tag release" $m/$tag;
  else
    hub release create -m "$m/$tag release" -t $commitsh $m/$tag;
  fi
done