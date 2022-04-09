package grpc

import (
	proto "github.com/go-micro/plugins/v4/config/source/grpc/proto"
	"go-micro.dev/v4/config/source"
)

type watcher struct {
	stream proto.Source_WatchClient
}

func newWatcher(stream proto.Source_WatchClient) (*watcher, error) {
	return &watcher{
		stream: stream,
	}, nil
}

func (w *watcher) Next() (*source.ChangeSet, error) {
	rsp, err := w.stream.Recv()
	if err != nil {
		return nil, err
	}
	return toChangeSet(rsp.ChangeSet), nil
}

func (w *watcher) Stop() error {
	return w.stream.CloseSend()
}
