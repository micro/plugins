package mucp

import (
	"time"

	proto "github.com/go-micro/plugins/v2/config/source/mucp/proto"
	"github.com/micro/go-micro/v2/config/source"
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
