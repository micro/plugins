//go:build windows
// +build windows

package windowseventlog

import "go-micro.dev/v5/logger"

type src struct{}
type eid struct{}

type Options struct {
	logger.Options
	Src string
	Eid uint32
}

func WithSrc(namesrc string) logger.Option {
	return logger.SetOption(src{}, namesrc)
}

func WithEid(neweid uint32) logger.Option {
	return logger.SetOption(eid{}, neweid)
}
