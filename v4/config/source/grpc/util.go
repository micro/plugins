package grpc

import (
	"time"

	proto "github.com/go-micro/plugins/v4/config/source/grpc/proto"
	"go-micro.dev/v4/config/source"
)

func toChangeSet(c *proto.ChangeSet) *source.ChangeSet {
	return &source.ChangeSet{
		Data:      c.Data,
		Checksum:  c.Checksum,
		Format:    c.Format,
		Timestamp: time.Unix(c.Timestamp, 0),
		Source:    c.Source,
	}
}
