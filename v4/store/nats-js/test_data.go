package natsjs

import "go-micro.dev/v4/store"

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
