package polaris

import (
	"time"

	"go-micro.dev/v5/registry"
)

type poWatcher struct {
	stop    chan bool
	timeout time.Duration
	opts    registry.WatchOptions
}

func newPoWatcher(timeout time.Duration, opts ...registry.WatchOption) (registry.Watcher, error) {
	w := &poWatcher{
		stop:    make(chan bool, 1),
		timeout: timeout,
	}

	for _, opt := range opts {
		opt(&w.opts)
	}

	return w, nil
}

// simulate for cache delete
// registry/cache.go watch,update ...
// otherwise, always called while deregister and not one service online.
func (pw *poWatcher) Next() (*registry.Result, error) {
	// just delete
	time.Sleep(pw.timeout)

	return &registry.Result{
		Action:  "delete",
		Service: &registry.Service{Name: pw.opts.Service},
	}, nil
}

func (pw *poWatcher) Stop() {
	select {
	case <-pw.stop:
		return
	default:
		close(pw.stop)
	}
}
