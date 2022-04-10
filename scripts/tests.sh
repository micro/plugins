#!/bin/bash -e

for d in $(find * -name 'go.mod'); do
  case $(dirname $d) in
    'v4/config/source/configmap'|'v4/config/source/vault'|'v4/events/redis'|'v4/logger/windowseventlog'|'v4/store/mysql'|'v4/store/redis'|'v4/sync/consul'|'v4/sync/etcd')
    echo skip $(dirname $d)
    ;;
    *)
    pushd $(dirname $d) >/dev/null
    go get
    go vet
    #go test -race -v
    go test -v
    popd >/dev/null
  esac
done
