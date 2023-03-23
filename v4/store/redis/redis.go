package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/store"
	"go-micro.dev/v4/util/cmd"
)

// DefaultDatabase is the namespace that the store
// will use if no namespace is provided.
var (
	DefaultDatabase = "micro"
	DefaultTable    = "micro"
)

type rkv struct {
	ctx     context.Context
	options store.Options
	Client  redis.UniversalClient
}

func init() {
	cmd.DefaultStores["redis"] = NewStore
}

func (r *rkv) Init(opts ...store.Option) error {
	for _, o := range opts {
		o(&r.options)
	}

	return r.configure()
}

func (r *rkv) Close() error {
	return r.Client.Close()
}

func (r *rkv) Read(key string, opts ...store.ReadOption) ([]*store.Record, error) {
	options := store.ReadOptions{}
	options.Table = r.options.Table

	for _, o := range opts {
		o(&options)
	}

	var keys []string

	rkey := fmt.Sprintf("%s%s", options.Table, key)
	// Handle Prefix
	// TODO suffix
	if options.Prefix {
		prefixKey := fmt.Sprintf("%s*", rkey)
		fkeys, err := r.Client.Keys(r.ctx, prefixKey).Result()
		if err != nil {
			return nil, err
		}
		// TODO Limit Offset

		keys = append(keys, fkeys...)
	} else {
		keys = []string{rkey}
	}

	records := make([]*store.Record, 0, len(keys))

	for _, rkey = range keys {
		val, err := r.Client.Get(r.ctx, rkey).Bytes()

		if err != nil && err == redis.Nil {
			return nil, store.ErrNotFound
		} else if err != nil {
			return nil, err
		}

		if val == nil {
			return nil, store.ErrNotFound
		}

		d, err := r.Client.TTL(r.ctx, rkey).Result()
		if err != nil {
			return nil, err
		}

		records = append(records, &store.Record{
			Key:    key,
			Value:  val,
			Expiry: d,
		})
	}

	return records, nil
}

func (r *rkv) Delete(key string, opts ...store.DeleteOption) error {
	options := store.DeleteOptions{}
	options.Table = r.options.Table

	for _, o := range opts {
		o(&options)
	}

	rkey := fmt.Sprintf("%s%s", options.Table, key)
	return r.Client.Del(r.ctx, rkey).Err()
}

func (r *rkv) Write(record *store.Record, opts ...store.WriteOption) error {
	options := store.WriteOptions{}
	options.Table = r.options.Table

	for _, o := range opts {
		o(&options)
	}

	rkey := fmt.Sprintf("%s%s", options.Table, record.Key)
	return r.Client.Set(r.ctx, rkey, record.Value, record.Expiry).Err()
}

func (r *rkv) List(opts ...store.ListOption) ([]string, error) {
	options := store.ListOptions{
		Table: r.options.Table,
	}

	for _, o := range opts {
		o(&options)
	}

	key := fmt.Sprintf("%s*%s", options.Prefix, options.Suffix)

	cursor := uint64(options.Offset)
	count := int64(options.Limit)
	var allKeys []string
	for {
		var err error
		var keys []string
		keys, cursor, err = r.Client.Scan(r.ctx, cursor, key, count).Result()
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	return allKeys, nil
}

func (r *rkv) Options() store.Options {
	return r.options
}

func (r *rkv) String() string {
	return "redis"
}

func NewStore(opts ...store.Option) store.Store {
	options := store.Options{
		Database: DefaultDatabase,
		Table:    DefaultTable,
		Logger:   logger.DefaultLogger,
	}

	for _, o := range opts {
		o(&options)
	}

	s := &rkv{
		ctx:     context.Background(),
		options: options,
	}

	if err := s.configure(); err != nil {
		s.options.Logger.Log(logger.ErrorLevel, "Error configuring store ", err)
	}

	return s
}

func (r *rkv) configure() error {
	if r.Client != nil {
		r.Client.Close()
	}
	r.Client = newUniversalClient(r.options)

	return nil
}
