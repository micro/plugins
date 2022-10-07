package polaris

import (
	"time"

	"go-micro.dev/v4/registry"
)

type poWatcher struct {
	stop    chan bool
	timeout time.Duration
	opts    registry.WatchOptions
}

func newPoWatcher(e *poRegistry, timeout time.Duration, opts ...registry.WatchOption) (registry.Watcher, error) {
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
// otherwise, always called while deregister and not one service online
func (ew *poWatcher) Next() (*registry.Result, error) {
	// just delete
	time.Sleep(ew.timeout)
	return &registry.Result{
		Action:  "delete",
		Service: &registry.Service{Name: ew.opts.Service},
	}, nil
}

func (ew *poWatcher) Stop() {
	select {
	case <-ew.stop:
		return
	default:
		close(ew.stop)
	}
}
