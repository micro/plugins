package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	consul "github.com/hashicorp/consul/api"
	"go-micro.dev/v4/registry"
)

type mockRegistry struct {
	body   []byte
	fn     func(r *http.Request) ([]byte, int, error)
	status int
	err    error
	url    string
}

func encodeData(obj interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(obj); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func newMockServer(l net.Listener, rgs ...*mockRegistry) error {
	mux := http.NewServeMux()
	for _, rg := range rgs {
		rgIn := rg
		mux.HandleFunc(rg.url, func(w http.ResponseWriter, r *http.Request) {
			body, status, err := rgIn.body, rgIn.status, rgIn.err
			if rg.fn != nil {
				body, status, err = rgIn.fn(r)
			}
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(status)
			w.Write(body)
		})
	}
	return http.Serve(l, mux)
}

func newConsulTestRegistry(chechkTTL time.Duration, r ...*mockRegistry) (*consulRegistry, func()) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		// blurgh?!!
		panic(err.Error())
	}
	cfg := consul.DefaultConfig()
	cfg.Address = l.Addr().String()

	go newMockServer(l, r...)

	var cr = &consulRegistry{
		config:      cfg,
		Address:     []string{cfg.Address},
		opts:        registry.Options{},
		register:    make(map[string]uint64),
		lastChecked: make(map[string]time.Time),
		queryOptions: &consul.QueryOptions{
			AllowStale: true,
		},
	}
	CheckTTL(time.Nanosecond)(&cr.opts)
	cr.Client()

	return cr, func() {
		l.Close()
	}
}

func newServiceList(svc []*consul.ServiceEntry) []byte {
	bts, _ := encodeData(svc)
	return bts
}

func TestConsul_GetService_WithError(t *testing.T) {
	cr, cl := newConsulTestRegistry(time.Second, &mockRegistry{
		err: errors.New("client-error"),
		url: "/v1/health/service/service-name",
	})
	defer cl()

	if _, err := cr.GetService("test-service"); err == nil {
		t.Fatalf("Expected error not to be `nil`")
	}
}

func TestConsul_GetService_WithHealthyServiceNodes(t *testing.T) {
	// warning is still seen as healthy, critical is not
	svcs := []*consul.ServiceEntry{
		newServiceEntry(
			"node-name-1", "node-address-1", "service-name", "v1.0.0",
			[]*consul.HealthCheck{
				newHealthCheck("node-name-1", "service-name", "passing"),
				newHealthCheck("node-name-1", "service-name", "warning"),
			},
		),
		newServiceEntry(
			"node-name-2", "node-address-2", "service-name", "v1.0.0",
			[]*consul.HealthCheck{
				newHealthCheck("node-name-2", "service-name", "passing"),
				newHealthCheck("node-name-2", "service-name", "warning"),
			},
		),
	}

	cr, cl := newConsulTestRegistry(time.Second, &mockRegistry{
		status: 200,
		body:   newServiceList(svcs),
		url:    "/v1/health/service/service-name",
	})
	defer cl()

	svc, err := cr.GetService("service-name")
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if exp, act := 1, len(svc); exp != act {
		t.Fatalf("Expected len of svc to be `%d`, got `%d`.", exp, act)
	}

	if exp, act := 2, len(svc[0].Nodes); exp != act {
		t.Fatalf("Expected len of nodes to be `%d`, got `%d`.", exp, act)
	}
}

func TestConsul_GetService_WithUnhealthyServiceNode(t *testing.T) {
	// warning is still seen as healthy, critical is not
	svcs := []*consul.ServiceEntry{
		newServiceEntry(
			"node-name-1", "node-address-1", "service-name", "v1.0.0",
			[]*consul.HealthCheck{
				newHealthCheck("node-name-1", "service-name", "passing"),
				newHealthCheck("node-name-1", "service-name", "warning"),
			},
		),
		newServiceEntry(
			"node-name-2", "node-address-2", "service-name", "v1.0.0",
			[]*consul.HealthCheck{
				newHealthCheck("node-name-2", "service-name", "passing"),
				newHealthCheck("node-name-2", "service-name", "critical"),
			},
		),
	}

	cr, cl := newConsulTestRegistry(time.Second, &mockRegistry{
		status: 200,
		body:   newServiceList(svcs),
		url:    "/v1/health/service/service-name",
	})
	defer cl()

	svc, err := cr.GetService("service-name")
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if exp, act := 1, len(svc); exp != act {
		t.Fatalf("Expected len of svc to be `%d`, got `%d`.", exp, act)
	}

	if exp, act := 1, len(svc[0].Nodes); exp != act {
		t.Fatalf("Expected len of nodes to be `%d`, got `%d`.", exp, act)
	}
}

func TestConsul_GetService_WithUnhealthyServiceNodes(t *testing.T) {
	// warning is still seen as healthy, critical is not
	svcs := []*consul.ServiceEntry{
		newServiceEntry(
			"node-name-1", "node-address-1", "service-name", "v1.0.0",
			[]*consul.HealthCheck{
				newHealthCheck("node-name-1", "service-name", "passing"),
				newHealthCheck("node-name-1", "service-name", "critical"),
			},
		),
		newServiceEntry(
			"node-name-2", "node-address-2", "service-name", "v1.0.0",
			[]*consul.HealthCheck{
				newHealthCheck("node-name-2", "service-name", "passing"),
				newHealthCheck("node-name-2", "service-name", "critical"),
			},
		),
	}

	cr, cl := newConsulTestRegistry(time.Second, &mockRegistry{
		status: 200,
		body:   newServiceList(svcs),
		url:    "/v1/health/service/service-name",
	})
	defer cl()

	svc, err := cr.GetService("service-name")
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if exp, act := 1, len(svc); exp != act {
		t.Fatalf("Expected len of svc to be `%d`, got `%d`.", exp, act)
	}

	if exp, act := 0, len(svc[0].Nodes); exp != act {
		t.Fatalf("Expected len of nodes to be `%d`, got `%d`.", exp, act)
	}
}

func TestConsul_TestRegistrer(t *testing.T) {
	registerCalled := 0
	cr, cl := newConsulTestRegistry(
		time.Second,
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				b, _ := io.ReadAll(r.Body)
				exp := `{"ID":"nodeId","Name":"service1","Tags":["v-789c010000ffff00000001"],` +
					`"Address":"address","Check":{"TTL":"1s","DeregisterCriticalServiceAfter":"1m5s"},"Checks":null}`
				body := strings.TrimSpace(string(b))
				if body != exp {
					t.Fatalf("Expected request to be %s`, got `%s`.", exp, body)
				}
				registerCalled++
				return []byte(`{"success"":true}`), 200, nil
			},
			url: "/v1/agent/service/register",
		},
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				b, _ := io.ReadAll(r.Body)
				exp := `{"Status":"passing","Output":""}`
				body := strings.TrimSpace(string(b))
				if body != exp {
					t.Fatalf("Expected request to be %s`, got `%s`.", exp, body)
				}
				return nil, 200, nil
			},
			url: "/v1/agent/check/update/service:nodeId",
		},
	)
	defer cl()

	service := &registry.Service{
		Name: "service1",
		Nodes: []*registry.Node{
			{
				Address: "address",
				Id:      "nodeId",
			},
		},
	}
	rOpts := []registry.RegisterOption{registry.RegisterTTL(time.Second)}
	err := cr.Register(service, rOpts...)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	err = cr.Register(service, rOpts...)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	if registerCalled >= 1 {
		t.Fatalf("Expected run time to be %d`, got `%d`.", 1, registerCalled)
	}
}

func TestConsul_TestRegistrerWithCheck(t *testing.T) {
	registerCalled := 0
	cr, cl := newConsulTestRegistry(
		time.Nanosecond,
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				b, _ := io.ReadAll(r.Body)
				exp := `{"ID":"nodeId","Name":"service1","Tags":["v-789c010000ffff00000001"],` +
					`"Address":"address","Check":{"TTL":"1s","DeregisterCriticalServiceAfter":"1m5s"},"Checks":null}`
				body := strings.TrimSpace(string(b))
				if body != exp {
					t.Fatalf("Expected request to be %s`, got `%s`.", exp, body)
				}
				registerCalled++
				return []byte(`{"success"":true}`), 200, nil
			},
			url: "/v1/agent/service/register",
		},
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				b, _ := io.ReadAll(r.Body)
				exp := `{"Status":"passing","Output":""}`
				body := strings.TrimSpace(string(b))
				if body != exp {
					t.Fatalf("Expected request to be %s`, got `%s`.", exp, body)
				}
				return nil, 200, nil
			},
			url: "/v1/agent/check/update/service:nodeId",
		},
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				b, _ := encodeData([]*consul.HealthCheck{
					newHealthCheck("nodeId", "service1", "passing"),
				})
				return b, 200, nil
			},
			url: "/v1/health/checks/service1",
		},
	)
	defer cl()

	service := &registry.Service{
		Name: "service1",
		Nodes: []*registry.Node{
			{
				Address: "address",
				Id:      "nodeId",
			},
		},
	}
	rOpts := []registry.RegisterOption{registry.RegisterTTL(time.Second)}
	err := cr.Register(service, rOpts...)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	err = cr.Register(service, rOpts...)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if registerCalled >= 1 {
		t.Fatalf("Expected run time to be %d`, got `%d`.", 1, registerCalled)
	}
}

func TestConsul_TestRegistrerWithFailedCheck(t *testing.T) {
	registerCalled := 0
	deregisterCalled := 0
	cr, cl := newConsulTestRegistry(
		time.Nanosecond,
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				b, _ := io.ReadAll(r.Body)
				exp := `{"ID":"nodeId","Name":"service1","Tags":["v-789c010000ffff00000001"],` +
					`"Address":"address","Check":{"TTL":"1s","DeregisterCriticalServiceAfter":"1m5s"},"Checks":null}`
				body := strings.TrimSpace(string(b))
				if body != exp {
					t.Fatalf("Expected request to be %s`, got `%s`.", exp, body)
				}
				registerCalled++
				return []byte(`{"success"":true}`), 200, nil
			},
			url: "/v1/agent/service/register",
		},
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				deregisterCalled++
				return []byte(`{"success"":true}`), 200, nil
			},
			url: "/v1/agent/service/deregister/nodeId",
		},
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				b, _ := io.ReadAll(r.Body)
				exp := `{"Status":"passing","Output":""}`
				body := strings.TrimSpace(string(b))
				if body != exp {
					t.Fatalf("Expected request to be %s`, got `%s`.", exp, body)
				}
				return nil, 200, nil
			},
			url: "/v1/agent/check/update/service:nodeId",
		},
		&mockRegistry{
			fn: func(r *http.Request) ([]byte, int, error) {
				b, _ := encodeData([]*consul.HealthCheck{
					newHealthCheck("nodeIdsdfsd", "service1", "passing"),
				})
				return b, 200, nil
			},
			url: "/v1/health/checks/service1",
		},
	)
	defer cl()

	service := &registry.Service{
		Name: "service1",
		Nodes: []*registry.Node{
			{
				Address: "address",
				Id:      "nodeId",
			},
		},
	}
	rOpts := []registry.RegisterOption{registry.RegisterTTL(time.Second)}
	err := cr.Register(service, rOpts...)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	err = cr.Register(service, rOpts...)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	if registerCalled >= 3 {
		t.Fatalf("Expected register run time to be %d`, got `%d`.", 2, registerCalled)
	}
	if deregisterCalled < 1 {
		t.Fatalf("Expected deregister run time to be %d`, got `%d`.", 1, deregisterCalled)
	}
}
