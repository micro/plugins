package etcd

import (
	"fmt"
	"strings"

	"go-micro.dev/v5/config/encoder"
	"go-micro.dev/v5/logger"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func makeEvMap(e encoder.Encoder, data map[string]interface{}, kv []*clientv3.Event, stripPrefix string) (map[string]interface{}, error) {
	if data == nil {
		data = make(map[string]interface{})
	}
	var err error
	for _, v := range kv {
		switch mvccpb.Event_EventType(v.Type) {
		case mvccpb.DELETE:
			data, _ = update(e, data, (*mvccpb.KeyValue)(v.Kv), "delete", stripPrefix)
		default:
			data, err = update(e, data, (*mvccpb.KeyValue)(v.Kv), "insert", stripPrefix)
			if err != nil {
				return nil, err
			}
		}
	}

	return data, nil
}

func makeMap(e encoder.Encoder, kv []*mvccpb.KeyValue, stripPrefix string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var err error
	for _, v := range kv {
		d, e := update(e, data, v, "put", stripPrefix)
		if e != nil {
			err = e
			logger.Errorf("etcd makeMap err %v", e)
		} else {
			data = d
		}
	}
	return data, err
}

func update(e encoder.Encoder, data map[string]interface{}, v *mvccpb.KeyValue, action, stripPrefix string) (map[string]interface{}, error) {
	// remove prefix if non empty, and ensure leading / is removed as well
	vkey := strings.TrimPrefix(strings.TrimPrefix(string(v.Key), stripPrefix), "/")
	// split on prefix
	haveSplit := strings.Contains(vkey, "/")
	keys := strings.Split(vkey, "/")

	var vals interface{}
	err := e.Decode(v.Value, &vals)
	if "delete" != action && err != nil {
		return data, fmt.Errorf("faild decode value. v.key: %s, error: %s", v.Key, err)
	}

	if !haveSplit && len(keys) == 1 {
		switch action {
		case "delete":
			data = make(map[string]interface{})
		default:
			v, ok := vals.(map[string]interface{})
			if ok {
				data = v
			}
		}
		return data, nil
	}

	// set data for first iteration
	kvals := data
	// iterate the keys and make maps
	for i, k := range keys {
		kval, ok := kvals[k].(map[string]interface{})
		if !ok {
			// create next map
			kval = make(map[string]interface{})
			// set it
			kvals[k] = kval
		}

		// last key: write vals
		if l := len(keys) - 1; i == l {
			switch action {
			case "delete":
				delete(kvals, k)
			default:
				kvals[k] = vals
			}
			break
		}

		// set kvals for next iterator
		kvals = kval
	}

	return data, nil
}
