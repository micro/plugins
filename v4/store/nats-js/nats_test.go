package natsjs

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go-micro.dev/v4/store"
)

type test struct {
	Record   *store.Record
	Database string
	Table    string
	WantErr  bool
}

var (
	table = []test{
		{
			Record: &store.Record{
				Key:   "One",
				Value: []byte("First value"),
			},
			WantErr: false,
		},
		{
			Record: &store.Record{
				Key:   "Two",
				Value: []byte("Second value"),
			},
			Table:   "prefix_test",
			WantErr: false,
		},
		{
			Record: &store.Record{
				Key:   "Third",
				Value: []byte("Third value"),
			},
			Database: "new-bucket",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "Four",
				Value: []byte("Fourth value"),
			},
			Database: "new-bucket",
			Table:    "prefix_test",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "Alex",
				Value: []byte("Some value"),
			},
			Database: "prefix-test",
			Table:    "names",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "Jones",
				Value: []byte("Some value"),
			},
			Database: "prefix-test",
			Table:    "names",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "Adrianna",
				Value: []byte("Some value"),
			},
			Database: "prefix-test",
			Table:    "names",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "MexicoCity",
				Value: []byte("Some value"),
			},
			Database: "prefix-test",
			Table:    "cities",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "HoustonCity",
				Value: []byte("Some value"),
			},
			Database: "prefix-test",
			Table:    "cities",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "ZurichCity",
				Value: []byte("Some value"),
			},
			Database: "prefix-test",
			Table:    "cities",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "Helsinki",
				Value: []byte("Some value"),
			},
			Database: "prefix-test",
			Table:    "cities",
			WantErr:  false,
		},
		{
			Record: &store.Record{
				Key:   "testKeytest",
				Value: []byte("Some value"),
			},
			Table:   "some_table",
			WantErr: false,
		},
		{
			Record: &store.Record{
				Key:   "testSecondtest",
				Value: []byte("Some value"),
			},
			Table:   "some_table",
			WantErr: false,
		},
		{
			Record: &store.Record{
				Key:   "lalala",
				Value: []byte("Some value"),
			},
			Table:   "some_table",
			WantErr: false,
		},
		{
			Record: &store.Record{
				Key:   "testAnothertest",
				Value: []byte("Some value"),
			},
			WantErr: false,
		},
	}
)

func setupTest(t *testing.T) store.Store {
	s := NewStore()
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNats(t *testing.T) {
	// Setup without calling Init on purpose
	s := NewStore()
	defer s.Close()
	t.Log("Testing:", s.String())

	basicTest(s, t)
}

func TestOptions(t *testing.T) {
	s := NewStore(
		store.Nodes(nats.DefaultURL),
		DefaultMemory(),

		// Having a non-default description will trigger nats.ErrStreamNameAlreadyInUse
		//  since the buckets have been created in previous tests with a different description.
		DefaultDescription("My fancy description"),

		// Option has no effect in this context, just to test setting the option
		JetStreamOptions(nats.PublishAsyncMaxPending(256)),

		// Sets a custom NATS client name, just to test the NatsOptions() func
		NatsOptions(nats.Options{Name: "Go NATS Store Plugin Tests Client"}),

		ObjectStoreOptions(&nats.ObjectStoreConfig{
			Bucket:      "TestBucketName",
			Description: "This bucket is not used",
			TTL:         5 * time.Minute,
			MaxBytes:    1024,
			Storage:     nats.MemoryStorage,
			Replicas:    1,
		}),
	)
	defer s.Close()

	basicTest(s, t)
}

func TestTTL(t *testing.T) {
	s := NewStore(
		DefaultTTL(100*time.Millisecond),

		// Since these buckets will be new they will have the new description
		DefaultDescription("My fancy description"),
	)

	// Use a uuid to make sure a new bucket is created
	id := uuid.New().String()
	for _, r := range table {
		if err := s.Write(r.Record, store.WriteTo(r.Database+id, r.Table)); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second)

	for _, r := range table {
		res, err := s.Read(r.Record.Key, store.ReadFrom(r.Database+id, r.Table))
		if err != nil {
			t.Fatal(err)
		}
		if len(res) > 0 {
			t.Fatal("Fetched record while it should have expired")
		}
	}
}

func TestMetaData(t *testing.T) {
	s := setupTest(t)
	defer s.Close()

	record := store.Record{
		Key:   "KeyOne",
		Value: []byte("Some value"),
		Metadata: map[string]interface{}{
			"meta-one": "val",
			"meta-two": 5,
		},
		Expiry: 0,
	}
	bucket := "meta-data-test"
	if err := s.Write(&record, store.WriteTo(bucket, "")); err != nil {
		t.Fatal(err)
	}

	r, err := s.Read(record.Key, store.ReadFrom(bucket, ""))
	if err != nil {
		t.Fatal(err)
	}
	if len(r) == 0 {
		t.Fatal("No results found")
	}

	m := r[0].Metadata
	if m["meta-one"].(string) != record.Metadata["meta-one"].(string) ||
		m["meta-two"].(float64) != float64(record.Metadata["meta-two"].(int)) {
		t.Fatalf("Metadata does not match: (%+v) != (%+v)", m, record.Metadata)
	}
}

func TestDelete(t *testing.T) {
	s := setupTest(t)

	for _, r := range table {
		if err := s.Write(r.Record, store.WriteTo(r.Database, r.Table)); err != nil {
			t.Fatal(err)
		}

		if err := s.Delete(r.Record.Key, store.DeleteFrom(r.Database, r.Table)); err != nil {
			t.Fatal(err)
		}

		res, err := s.Read(r.Record.Key, store.ReadFrom(r.Database, r.Table))
		if err != nil {
			t.Fatal(err)
		}
		if len(res) > 0 {
			t.Fatalf("Failed to delete %s from %s %s proery", r.Record.Key, r.Database, r.Table)
		}
		t.Logf("Test %s passed", r.Record.Key)
	}
}

func TestList(t *testing.T) {
	s := setupTest(t)
	defer s.Close()

	for _, r := range table {
		if err := s.Write(r.Record, store.WriteTo(r.Database, r.Table)); err != nil {
			t.Fatal(err)
		}
	}

	l := []struct {
		Database string
		Table    string
		Length   int
		Prefix   string
		Suffix   string
		Offset   int
		Limit    int
	}{
		{Length: 6},
		{Database: "prefix-test", Length: 7},
		{Database: "prefix-test", Offset: 2, Length: 5},
		{Database: "prefix-test", Offset: 2, Limit: 3, Length: 3},
		{Database: "prefix-test", Table: "names", Length: 3},
		{Database: "prefix-test", Table: "cities", Length: 4},
		{Database: "prefix-test", Table: "cities", Suffix: "City", Length: 3},
		{Database: "prefix-test", Table: "cities", Suffix: "City", Limit: 2, Length: 2},
		{Database: "prefix-test", Table: "cities", Suffix: "City", Offset: 1, Length: 2},
		{Prefix: "test", Length: 1},
		{Table: "some_table", Prefix: "test", Suffix: "test", Length: 2},
	}

	for i, entry := range l {
		// Test listing keys
		keys, err := s.List(
			store.ListFrom(entry.Database, entry.Table),
			store.ListPrefix(entry.Prefix),
			store.ListSuffix(entry.Suffix),
			store.ListOffset(uint(entry.Offset)),
			store.ListLimit(uint(entry.Limit)),
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(keys) != entry.Length {
			t.Fatalf("Length of returned keys is invalid for test %d - %+v (%d)", i+1, entry, len(keys))
		}

		// Test reading keys
		if entry.Prefix != "" || entry.Suffix != "" {
			var key string
			options := []store.ReadOption{
				store.ReadFrom(entry.Database, entry.Table),
				store.ReadLimit(uint(entry.Limit)),
				store.ReadOffset(uint(entry.Offset)),
			}
			if entry.Prefix != "" {
				key = entry.Prefix
				options = append(options, store.ReadPrefix())
			}
			if entry.Suffix != "" {
				key = entry.Suffix
				options = append(options, store.ReadSuffix())
			}
			r, err := s.Read(key, options...)
			if err != nil {
				t.Fatal(err)
			}
			if len(r) != entry.Length {
				t.Fatalf("Length of read keys is invalid for test %d - %+v (%d)", i+1, entry, len(r))
			}
		}
	}
}

func TestDeleteBucket(t *testing.T) {
	s := setupTest(t)
	defer s.Close()

	for _, r := range table {
		if err := s.Write(r.Record, store.WriteTo(r.Database, r.Table)); err != nil {
			t.Fatal(err)
		}
	}

	bucket := "prefix-test"
	if err := s.Delete(bucket, DeleteBucket()); err != nil {
		t.Fatal(err)
	}

	keys, err := s.List(store.ListFrom(bucket, ""))
	if err != ErrBucketNotFound && err != nil {
		t.Fatalf("Failed to delete bucket: %v", err)
	}

	if len(keys) > 0 {
		t.Fatal("Length of key list should be 0 after bucket deletion")
	}

	r, err := s.Read("", store.ReadPrefix(), store.ReadFrom(bucket, ""))
	if err != ErrBucketNotFound && err != nil {
		t.Fatalf("Failed to delete bucket: %v", err)
	}
	if len(r) > 0 {
		t.Fatal("Length of record list should be 0 after bucket deletion", len(r))
	}

}

func basicTest(s store.Store, t *testing.T) {
	for _, test := range table {
		if err := s.Write(test.Record, store.WriteTo(test.Database, test.Table)); err != nil {
			t.Fatal(err)
		}
		r, err := s.Read(test.Record.Key, store.ReadFrom(test.Database, test.Table))
		if err != nil {
			t.Fatal(err)
		}
		if len(r) == 0 {
			t.Fatalf("No results found for %s (%s) %s", test.Record.Key, test.Database, test.Table)
		}

		key := test.Record.Key
		val1 := string(test.Record.Value)

		key2 := r[0].Key
		val2 := string(r[0].Value)
		if val1 != val2 {
			t.Fatalf("Value not equal for (%s: %s) != (%s: %s)", key, val1, key2, val2)
		}
		t.Logf("Test %s passed", test.Record.Key)
	}
}
