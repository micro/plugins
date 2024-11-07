package polaris_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"go-micro.dev/v5/registry"

	"github.com/google/uuid"

	"github.com/micro/plugins/v5/registry/polaris"
)

const (
	defaultToken     = "nu/0WRA4EqSR1FagrjRj0fZwPXuGlMpX+zCuWu4uMqy8xr1vRjisSbA25aAC3mtU8MeeRsKhQiDAynUR09I="
	defaultNamespace = "default"
	defaultTimeout   = time.Second * 30
	defaultAddr      = "127.0.0.1:8091"
	addrEnv          = "POLARIS_ADDR"
)

func newRegistry() registry.Registry {
	addr := defaultAddr
	if env := os.Getenv("POLARIS_ADDR"); len(env) > 0 {
		addr = env
	}

	reg := polaris.NewRegistry(
		registry.Addrs(addr),
		registry.Timeout(defaultTimeout),
		polaris.GetOneInstance(false),
		polaris.NameSpace(defaultNamespace),
		polaris.ServerToken(defaultToken),
	)

	return reg
}

func genService(name, version, addr string) registry.Service {
	service := registry.Service{
		Name:     name,
		Version:  version,
		Metadata: map[string]string{},
		Endpoints: []*registry.Endpoint{
			{Name: "call"},
		},
		Nodes: []*registry.Node{
			{
				Id:      name + "-" + version + "-" + uuid.NewString(),
				Address: addr,
			},
		},
	}

	return service
}

func TestRegister(t *testing.T) {
	// Wait for Deregister
	// If Deregister be interrupt ,will effect other test with same address
	defer time.Sleep(time.Second * 10)

	// Polaris reg wait time
	regWait := time.Second * 8

	t.Log("Creating new registry")
	reg := newRegistry()

	{
		// Register Service
		t.Log("Registering first service")
		srv := genService("test", "v1", "127.0.0.1:6000")
		assertNoError(t, reg.Register(&srv, registry.RegisterTTL(time.Second*30)))
		defer func(srv *registry.Service) {
			if err := reg.Deregister(srv); err != nil {
				t.Fatalf("Failed to Deregister: %v", err)
			}
		}(&srv)

		time.Sleep(regWait)

		t.Log("Fetching first service")
		services, err := reg.GetService("test")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
			t.SkipNow()
			return
		}

		for i, item := range services {
			t.Logf("ID: %d, Name: %s, Version: %v, Node Count %v, First Node Address %v, First Node ID: %v\n",
				i, item.Name, item.Version, len(item.Nodes), item.Nodes[0].Address, item.Nodes[0].Id)
		}

		if len(services) != 1 {
			t.Errorf("expected %v, got %v", 1, len(services))
		}
	}

	{
		// Register Service
		t.Log("Registering second service")
		srv := genService("test", "v1", "127.0.0.1:7000")
		assertNoError(t, reg.Register(&srv, registry.RegisterTTL(time.Second*30)))
		defer func(srv *registry.Service) {
			if err := reg.Deregister(srv); err != nil {
				t.Fatalf("Failed to Deregister: %v", err)
			}
		}(&srv)

		time.Sleep(regWait)

		t.Log("Fetching second service")
		services, err := reg.GetService("test")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
			t.SkipNow()
			return
		}

		for i, item := range services {
			t.Logf("ID: %d, Name: %s, Version: %v, Node Count %v, First Node Address %v, First Node ID: %v\n",
				i, item.Name, item.Version, len(item.Nodes), item.Nodes[0].Address, item.Nodes[0].Id)
		}
		if len(services) != 1 {
			t.Errorf("expected %v, got %v", 1, len(services))
		} else if len(services[0].Nodes) != 2 {
			t.Errorf("expected %v, got %v", 2, len(services[0].Nodes))
		}
	}

	{
		// Register Service
		t.Log("Registering third service")
		srv := genService("test", "v2", "127.0.0.1:8000")
		assertNoError(t, reg.Register(&srv, registry.RegisterTTL(time.Second*30)))
		defer func(srv *registry.Service) {
			if err := reg.Deregister(srv); err != nil {
				t.Fatalf("Failed to Deregister: %v", err)
			}
		}(&srv)

		time.Sleep(regWait)

		t.Log("Fetching third service")
		services, err := reg.GetService("test")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
			t.SkipNow()
			return
		}

		for i, item := range services {
			t.Logf("ID: %d, Name: %s, Version: %v, Node Count %v, First Node Address %v, First Node ID: %v\n",
				i, item.Name, item.Version, len(item.Nodes), item.Nodes[0].Address, item.Nodes[0].Id)
		}

		if len(services) != 2 {
			t.Errorf("expected %v, got %v", 2, len(services))
		}
	}
}

func TestDeregister(t *testing.T) {
	// Wait for Deregister
	// If Deregister be interrupt ,will effect other test with same address
	defer time.Sleep(time.Second * 10)

	// Polaris reg wait time
	regWait := time.Second * 8

	t.Log("Creating new registry")
	reg := newRegistry()
	service1 := genService("test-deregister", "v1", "127.0.0.1:6100")
	service2 := genService("test-deregister", "v2", "127.0.0.1:6101")

	// Register 1
	t.Log("Register nr. 1")
	assertNoError(t, reg.Register(&service1, registry.RegisterTTL(time.Second*30)))
	time.Sleep(regWait)
	services, err := reg.GetService(service1.Name)
	assertNoError(t, err)
	assertSrvLen(t, 1, services)

	// Register 2
	t.Log("Register nr. 2")
	assertNoError(t, reg.Register(&service2, registry.RegisterTTL(time.Second*30)))
	time.Sleep(regWait)
	services, err = reg.GetService(service2.Name)
	assertNoError(t, err)
	assertSrvLen(t, 2, services)

	// Deregister 1
	t.Log("Deregister nr. 1")
	assertNoError(t, reg.Deregister(&service1))
	time.Sleep(regWait * 2)
	services, err = reg.GetService(service1.Name)
	assertNoError(t, err)
	assertSrvLen(t, 1, services)

	// Deregister 2
	t.Log("Deregister nr. 2")
	assertNoError(t, reg.Deregister(&service2))
	time.Sleep(regWait * 2)
	servicesList, err := reg.GetService(service2.Name)
	if !errors.Is(err, registry.ErrNotFound) {
		t.Error("expected err got nil")
	}
	assertSrvLen(t, 0, servicesList)
}

func BenchmarkGetService(b *testing.B) {
	b.StopTimer()

	// time.Sleep(time.Second * 10)

	b.Log("Registering test service")
	reg := newRegistry()

	// If the service name is constant here, the benchmark will often fail with
	// err len == 0, as Polaris marks the service as Unhealthy
	srvName := uuid.NewString()

	srv := genService(srvName, "v1", "127.0.0.1:6200")
	assertNoError(b, reg.Register(&srv))
	defer func(srv *registry.Service) {
		if err := reg.Deregister(srv); err != nil {
			b.Fatalf("Failed to Deregister: %v", err)
		}
	}(&srv)

	// Give Polaris some time to register the service.
	// One second should work, but some safety margin added.
	time.Sleep(time.Second * 5)

	b.Log("Starting benchmark")
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		services, err := reg.GetService(srvName)
		assertNoError(b, err)
		assertSrvLen(b, 1, services)
		assertEqual(b, srvName, services[0].Name)
	}
}
