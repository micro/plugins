package natsjs

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	nserver "github.com/nats-io/nats-server/v2/server"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go-micro.dev/v4/store"
)

func testSetup(ctx context.Context, t *testing.T, opts ...store.Option) store.Store {
	addr := startNatsServer(ctx, t)

	opts = append(opts, store.Nodes(addr))
	s := NewStore(opts...)

	if err := s.Init(); err != nil {
		t.Fatal(errors.Wrap(err, "Store initialization failed"))
	}

	go func() {
		<-ctx.Done()
		s.Close()
	}()

	return s
}

func startNatsServer(ctx context.Context, t *testing.T) string {
	natsAddr := getFreeLocalhostAddress()
	natsPort, _ := strconv.Atoi(strings.Split(natsAddr, ":")[1])

	clusterName := "gomicro-store-test-cluster"

	// start the NATS with JetStream server
	go natsServer(ctx,
		t,
		&nserver.Options{
			Host: strings.Split(natsAddr, ":")[0],
			Port: natsPort,
			Cluster: nserver.ClusterOpts{
				Name: clusterName,
			},
		},
	)

	time.Sleep(1 * time.Second)

	return natsAddr
}

func getFreeLocalhostAddress() string {
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	defer l.Close()
	return l.Addr().String()
}

func natsServer(ctx context.Context, t *testing.T, opts *nserver.Options) {
	server, err := nserver.NewServer(
		opts,
	)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	server.SetLoggerV2(
		NewLogWrapper(),
		false, false, false,
	)

	// first start NATS
	go server.Start()

	tmpdir := t.TempDir()
	natsdir := filepath.Join(tmpdir, "nats-js")

	jsConf := &nserver.JetStreamConfig{
		StoreDir: natsdir,
	}

	// This fixes some issues where tests fail because directory cleanup fails
	t.Cleanup(func() {
		contents, _ := filepath.Glob(natsdir + "/*")
		for _, item := range contents {
			os.RemoveAll(item)
		}
		os.RemoveAll(natsdir)
	})

	// second start JetStream
	err = server.EnableJetStream(jsConf)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	<-ctx.Done()

	server.Shutdown()
}

func NewLogWrapper() *LogWrapper {
	return &LogWrapper{}
}

type LogWrapper struct {
}

// Noticef logs a notice statement
func (l *LogWrapper) Noticef(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

// Warnf logs a warning statement
func (l *LogWrapper) Warnf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

// Fatalf logs a fatal statement
func (l *LogWrapper) Fatalf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

// Errorf logs an error statement
func (l *LogWrapper) Errorf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

// Debugf logs a debug statement
func (l *LogWrapper) Debugf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

// Tracef logs a trace statement
func (l *LogWrapper) Tracef(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}
