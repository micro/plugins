package redis

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"go-micro.dev/v5/store"
)

func Test_rkv_configure(t *testing.T) {
	type fields struct {
		options store.Options
	}
	type wantValues struct {
		username string
		password string
		address  string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		want    wantValues
	}{
		{name: "No Url", fields: fields{options: store.Options{}},
			wantErr: false, want: wantValues{
				username: "",
				password: "",
				address:  "127.0.0.1:6379",
			}},
		{name: "legacy Url", fields: fields{options: store.Options{Nodes: []string{"127.0.0.1:6379"}}},
			wantErr: false, want: wantValues{
				username: "",
				password: "",
				address:  "127.0.0.1:6379",
			}},
		{name: "New Url", fields: fields{options: store.Options{Nodes: []string{"redis://127.0.0.1:6379"}}},
			wantErr: false, want: wantValues{
				username: "",
				password: "",
				address:  "127.0.0.1:6379",
			}},
		{name: "Url with Pwd", fields: fields{options: store.Options{Nodes: []string{"redis://:password@redis:6379"}}},
			wantErr: false, want: wantValues{
				username: "",
				password: "password",
				address:  "redis:6379",
			}},
		{name: "Url with username and Pwd", fields: fields{
			options: store.Options{Nodes: []string{"redis://username:password@redis:6379"}}},
			wantErr: false, want: wantValues{
				username: "username",
				password: "password",
				address:  "redis:6379",
			}},

		{name: "Sentinel Failover client", fields: fields{
			options: store.Options{
				Nodes: []string{"127.0.0.1:6379", "127.0.0.1:6380"},
				Context: context.WithValue(
					context.TODO(), redisOptionsContextKey{},
					redis.UniversalOptions{MasterName: "master-name"}),
			}},
			wantErr: false, want: wantValues{
				username: "",
				password: "",
				address:  "FailoverClient", // <- Placeholder set by NewFailoverClient
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &rkv{
				options: tt.fields.options,
			}
			err := r.configure()
			if (err != nil) != tt.wantErr {
				t.Errorf("configure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			client, ok := r.Client.(*redis.Client)
			if !ok {
				t.Errorf("configure() expect a *redis.Client")
				return
			}
			if client.Options().Addr != tt.want.address {
				t.Errorf("configure() Address = %v, want address %v", client.Options().Addr, tt.want.address)
			}
			if client.Options().Password != tt.want.password {
				t.Errorf("configure() password = %v, want password %v", client.Options().Password, tt.want.password)
			}
			if client.Options().Username != tt.want.username {
				t.Errorf("configure() username = %v, want username %v", client.Options().Username, tt.want.username)
			}
		})
	}
}

func Test_rkv_configure_cluster(t *testing.T) {
	type fields struct {
		options store.Options
	}
	type wantValues struct {
		username string
		password string
		addrs    []string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		want    wantValues
	}{
		{name: "Nodes", fields: fields{options: store.Options{Nodes: []string{"127.0.0.1:6379", "127.0.0.1:6380"}}},
			wantErr: false, want: wantValues{
				username: "",
				password: "",
				addrs:    []string{"127.0.0.1:6379", "127.0.0.1:6380"},
			}},
		{name: "Nodes with redis options", fields: fields{
			options: store.Options{
				Nodes: []string{"127.0.0.1:6379", "127.0.0.1:6380"},
				Context: context.WithValue(
					context.TODO(), redisOptionsContextKey{},
					redis.UniversalOptions{}),
			}},
			wantErr: false, want: wantValues{
				username: "",
				password: "",
				addrs:    []string{"127.0.0.1:6379", "127.0.0.1:6380"},
			}},
		{name: "Nodes in redis options", fields: fields{
			options: store.Options{
				Nodes: []string{"127.0.0.1:6379", "127.0.0.1:6380"}, // <- ignored
				Context: context.WithValue(
					context.TODO(), redisOptionsContextKey{},
					redis.UniversalOptions{Addrs: []string{"127.0.0.1:6381", "127.0.0.1:6382"}}),
			}},
			wantErr: false, want: wantValues{
				username: "",
				password: "",
				addrs:    []string{"127.0.0.1:6381", "127.0.0.1:6382"},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &rkv{
				options: tt.fields.options,
			}
			err := r.configure()
			if (err != nil) != tt.wantErr {
				t.Errorf("configure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			client, ok := r.Client.(*redis.ClusterClient)
			if !ok {
				t.Errorf("configure() expect a *redis.ClusterClient")
				return
			}
			if !reflect.DeepEqual(client.Options().Addrs, tt.want.addrs) {
				t.Errorf("configure() Addrs = %v, want addrs %v", client.Options().Addrs, tt.want.addrs)
			}
			if client.Options().Password != tt.want.password {
				t.Errorf("configure() password = %v, want password %v", client.Options().Password, tt.want.password)
			}
			if client.Options().Username != tt.want.username {
				t.Errorf("configure() username = %v, want username %v", client.Options().Username, tt.want.username)
			}
		})
	}
}

func Test_Store(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if len(url) == 0 {
		t.Skip("REDIS_URL not defined")
	}

	r := NewStore(store.Nodes(url))

	key := "myTest"
	rec := store.Record{
		Key:    key,
		Value:  []byte("myValue"),
		Expiry: 2 * time.Minute,
	}

	err := r.Write(&rec)
	if err != nil {
		t.Errorf("Write Erroe. Error: %v", err)
	}

	rec1, err := r.Read(key)
	if err != nil {
		t.Errorf("Read Error. Error: %v\n", err)
	}

	keys, err := r.List()
	if err != nil {
		t.Errorf("listing error %v\n", err)
	}
	if len(keys) < 1 {
		t.Errorf("not enough keys\n")
	}

	err = r.Delete(rec1[0].Key)
	if err != nil {
		t.Errorf("Delete error %v\n", err)
	}

	keys, err = r.List()
	if err != nil {
		t.Errorf("listing error %v\n", err)
	}
	if len(keys) > 0 {
		t.Errorf("too many keys\n")
	}
}
