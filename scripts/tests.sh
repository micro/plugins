#!/bin/bash -e

for d in $(find 'v3' -name 'go.mod'); do
  case $(dirname $d) in
    'v3/config/source/configmap'|'v3/config/source/vault'|'v3/events/redis'|'v3/logger/apex'|'v3/logger/windowseventlog'|'v3/store/mysql'|'v3/store/redis'|'v3/store/memory'|'v3/store/file'|'v3/sync/consul'|'v3/sync/etcd')
    echo SKIP $(dirname $d)
    ;;
    *)
    pushd $(dirname $d) >/dev/null
    go vet
    go test
    popd >/dev/null
  esac
done

for d in $(find 'v4' -name 'go.mod'); do
  case $(dirname $d) in
    'v4/config/source/configmap'|'v4/config/source/vault'|'v4/events/redis'|'v4/events/natsjs'|'v4/logger/windowseventlog'|'v4/store/mysql'|'v4/store/redis'|'v4/sync/consul'|'v4/sync/etcd')
    echo SKIP $(dirname $d)
    ;;
    *)
    pushd $(dirname $d) >/dev/null
    go vet
    go test # -race
    popd >/dev/null
  esac
done
