#!/bin/bash -e

version=$1

if [ "x$version" = "x" ]; then
  echo "must specify version"
  exit 1;
fi

for d in $(find $version -name 'go.mod'); do
  case $(dirname $d) in
    'v2/agent/command/animate'|'v2/broker/googlepubsub'|'v2/config/source/configmap'|'v2/config/source/vault'|'v2/events/redis'|'v2/logger/windowseventlog'|'v2/logger/zerolog'|'v2/registry/kubernetes'|'v2/store/mysql'|'v2/store/redis'|'v2/store/memory'|'v2/store/file'|'v2/sync/consul'|'v2/sync/etcd')
    echo SKIP $(dirname $d)
    ;;
    'v3/config/source/configmap'|'v3/config/source/vault'|'v3/events/redis'|'v3/logger/apex'|'v3/logger/windowseventlog'|'v3/store/mysql'|'v3/store/redis'|'v3/store/memory'|'v3/store/file'|'v3/sync/consul'|'v3/sync/etcd')
    echo SKIP $(dirname $d)
    ;;
    'v4/config/source/configmap'|'v4/config/source/vault'|'v4/events/redis'|'v4/events/natsjs'|'v4/logger/windowseventlog'|'v4/store/mysql'|'v4/store/redis'|'v4/sync/consul'|'v4/sync/etcd')
    echo SKIP $(dirname $d)
    ;;
    *)
    pushd $(dirname $d) >/dev/null
    go vet
    go test
    popd >/dev/null
  esac
done
