// Package polaris provides an etcd service registry
// https://github.com/polarismesh/polaris
package polaris

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/model"

	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/cmd"
)

var (
	prefix      = "/micro/registry/"
	defaultAddr = "127.0.0.1:8091"

	// DefaultTimeout is the default registry timeout.
	DefaultTimeout = time.Second * 5

	// ErrNoNodes is returned when the lenth of the node slice is zero.
	ErrNoNodes = errors.New("no nodes provided")
	// ErrNoMetadata is returned when metadata failed to fetch.
	ErrNoMetadata = errors.New("fail to GetMetadata")
)

type polarisRegistry struct {
	sync.RWMutex

	options registry.Options

	// only get one instance for use polaris's loadbalance
	getOneInstance bool
	register       map[string]string
	namespace      string
	serverToken    string

	provider api.ProviderAPI
	consumer api.ConsumerAPI
}

func init() {
	cmd.DefaultRegistries["polaris"] = NewRegistry
}

// NewRegistry creates a new Polaris registry.
func NewRegistry(opts ...registry.Option) registry.Registry {
	polaris := polarisRegistry{
		options:        *registry.NewOptions(opts...),
		register:       make(map[string]string),
		getOneInstance: false,
	}

	polaris.options.Timeout = DefaultTimeout

	if token := os.Getenv("POLARIS_TOKEN"); token != "" {
		opts = append(opts, ServerToken(token))
	}

	if ns := os.Getenv("POLARIS_NAMESPACE"); ns != "" {
		opts = append(opts, NameSpace(ns))
	}

	if address := os.Getenv("MICRO_REGISTRY_ADDRESS"); len(address) > 0 {
		opts = append(opts, registry.Addrs(address))
	}

	if err := polaris.configure(opts...); err != nil {
		polaris.Options().Logger.Logf(log.ErrorLevel, "failed to create Polaris registry: %v", err)
	}

	return &polaris
}

func (p *polarisRegistry) Init(opts ...registry.Option) error {
	return p.configure(opts...)
}

func (p *polarisRegistry) configure(opts ...registry.Option) error {
	for _, o := range opts {
		o(&p.options)
	}

	if p.options.Context != nil {
		if ns, ok := p.options.Context.Value(nameSpaceKey{}).(string); ok {
			p.namespace = ns
		}

		if token, ok := p.options.Context.Value(serverTokenKey{}).(string); ok {
			p.serverToken = token
		}

		if flag, ok := p.options.Context.Value(getOneInstanceKey{}).(bool); ok {
			p.getOneInstance = flag
		}
	}

	addr := defaultAddr

	for _, a := range p.Options().Addrs {
		if a != "" {
			addr = a
		}
	}

	consumer, err := api.NewConsumerAPIByAddress(addr)
	if err != nil {
		return err
	}

	provider, err := api.NewProviderAPIByAddress(addr)
	if err != nil {
		return err
	}

	p.consumer = consumer
	p.provider = provider

	return nil
}

func (p *polarisRegistry) registerNode(service *registry.Service, node *registry.Node,
	opts ...registry.RegisterOption) error {
	nodeSrv := registry.Service{
		Name:      service.Name,
		Version:   service.Version,
		Metadata:  service.Metadata,
		Endpoints: service.Endpoints,
		Nodes:     []*registry.Node{node},
	}

	addrs := strings.Split(node.Address, ":")
	if len(addrs) != 2 {
		return fmt.Errorf("fail to register instance, node.Address invalid: %s", node.Address)
	}

	host := addrs[0]

	port, err := strconv.Atoi(addrs[1])
	if err != nil {
		return err
	}

	retryCount := 3

	p.Lock()
	defer p.Unlock()

	// Check if node already registered
	if id := p.getInstance(node.Id); id != "" {
		req := &api.InstanceHeartbeatRequest{
			InstanceHeartbeatRequest: model.InstanceHeartbeatRequest{
				Service:      service.Name,
				Namespace:    p.namespace,
				Host:         host,
				Port:         port,
				ServiceToken: p.serverToken,
				RetryCount:   &retryCount,
				InstanceID:   id,
			},
		}

		if err = p.provider.Heartbeat(req); err != nil {
			return errors.Wrapf(err, "fail to heartbeat instance %+v", req)
		}

		return nil
	}

	var options registry.RegisterOptions

	for _, o := range opts {
		o(&options)
	}

	version := service.Version

	req := api.InstanceRegisterRequest{
		InstanceRegisterRequest: model.InstanceRegisterRequest{
			Service:      service.Name,
			Version:      &version,
			Namespace:    p.namespace,
			Host:         host,
			Port:         port,
			ServiceToken: p.serverToken,
			RetryCount:   &retryCount,
		},
	}

	req.SetTTL(int(options.TTL.Seconds()))

	b, err := json.Marshal(&nodeSrv)
	if err != nil {
		return err
	}

	req.Metadata = map[string]string{
		"node_path":     nodePath(nodeSrv.Name, node.Id),
		"Micro-Service": string(b),
	}

	resp, err := p.provider.Register(&req)
	if err != nil {
		return errors.Wrapf(err, "fail to register instance, err is %+v", req)
	}

	p.addInstance(node.Id, resp.InstanceID)

	return nil
}

// Deregister will deregister a node.
func (p *polarisRegistry) Deregister(s *registry.Service, opts ...registry.DeregisterOption) error {
	if len(s.Nodes) != 1 {
		return ErrNoNodes
	}

	p.Lock()
	defer p.Unlock()

	for _, node := range s.Nodes {
		addrs := strings.Split(node.Address, ":")
		if len(addrs) != 2 {
			return fmt.Errorf("fail to deregister instance, node.Address invalid %s", node.Address)
		}

		host := addrs[0]

		port, err := strconv.Atoi(addrs[1])
		if err != nil {
			return err
		}

		timeout := p.options.Timeout
		retryCount := 3

		req := api.InstanceDeRegisterRequest{
			InstanceDeRegisterRequest: model.InstanceDeRegisterRequest{
				Service:      s.Name,
				Namespace:    p.namespace,
				Host:         host,
				Port:         port,
				ServiceToken: p.serverToken,
				Timeout:      &timeout,
				RetryCount:   &retryCount,
			},
		}

		p.delInstance(node.Id)

		if err := p.provider.Deregister(&req); err != nil {
			return errors.Wrap(err, "fail to deregister instance")
		}
	}

	return nil
}

// Register will register a node.
func (p *polarisRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	var gerr error

	// register each node individually
	for _, node := range s.Nodes {
		if err := p.registerNode(s, node, opts...); err != nil {
			gerr = err
		}
	}

	return gerr
}

// GetService will fetch the service list.
func (p *polarisRegistry) GetService(name string, opts ...registry.GetOption) ([]*registry.Service, error) {
	logger := p.options.Logger
	timeout := p.options.Timeout

	retryCount := 3

	type getInstancer interface {
		GetInstances() []model.Instance
	}

	var (
		err     error
		insResp getInstancer
	)

	// DiscoverEchoServer
	if !p.getOneInstance {
		req := api.GetInstancesRequest{
			GetInstancesRequest: model.GetInstancesRequest{
				Service:    name,
				Namespace:  p.namespace,
				Timeout:    &timeout,
				RetryCount: &retryCount,
			},
		}

		insResp, err = p.consumer.GetInstances(&req)
		if err != nil {
			return nil, errors.Wrap(err, "fail to GetInstances")
		}
	} else {
		req := api.GetOneInstanceRequest{
			GetOneInstanceRequest: model.GetOneInstanceRequest{
				Service:    name,
				Namespace:  p.namespace,
				Timeout:    &timeout,
				RetryCount: &retryCount,
			},
		}

		insResp, err = p.consumer.GetOneInstance(&req)
		if err != nil {
			return nil, errors.Wrap(err, "fail to GetOneInstance")
		}
	}

	inss := insResp.GetInstances()
	if len(inss) == 0 {
		return []*registry.Service{}, registry.ErrNotFound
	}

	serviceMap := make(map[string]*registry.Service)

	for _, n := range inss {
		m := n.GetMetadata()
		if m == nil {
			return nil, ErrNoMetadata
		}

		microService := m["Micro-Service"]

		var service *registry.Service
		if err := json.Unmarshal([]byte(microService), &service); err != nil {
			return nil, err
		}

		if service != nil {
			if !n.IsHealthy() {
				msg := fmt.Sprintf("{Name: %s, Version: %v, Node Count: %v}", service.Name, service.Version, len(service.Nodes))
				logger.Logf(log.WarnLevel, "Service not healthy according to Polaris: %v", msg)

				continue
			}

			if n.IsIsolated() {
				msg := fmt.Sprintf("{Name: %s, Version: %v, Node Count: %v}", service.Name, service.Version, len(service.Nodes))
				logger.Logf(log.WarnLevel, "Service is isolated according to Polaris: %v", msg)

				continue
			}

			s, ok := serviceMap[service.Version]
			if !ok {
				s = &registry.Service{
					Name:      service.Name,
					Version:   service.Version,
					Metadata:  service.Metadata,
					Endpoints: service.Endpoints,
				}
				serviceMap[s.Version] = s
			}

			s.Nodes = append(s.Nodes, service.Nodes...)
		}
	}

	services := make([]*registry.Service, 0, len(serviceMap))

	for _, service := range serviceMap {
		services = append(services, service)
	}

	return services, nil
}

func (p *polarisRegistry) ListServices(opts ...registry.ListOption) ([]*registry.Service, error) {
	services := make([]*registry.Service, 0)
	return services, errors.New("not support")
}

func (p *polarisRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	return newPoWatcher(p.options.Timeout, opts...)
}

// String returns the Polaris plugin name.
func (p *polarisRegistry) String() string {
	return "polaris"
}

func nodePath(s, id string) string {
	service := strings.ReplaceAll(s, "/", "-")
	node := strings.ReplaceAll(id, "/", "-")

	return path.Join(prefix, service, node)
}

func (p *polarisRegistry) addInstance(nodeID, id string) {
	p.register[nodeID] = id
}

func (p *polarisRegistry) getInstance(nodeID string) string {
	if id, ok := p.register[nodeID]; ok {
		return id
	}

	return ""
}

func (p *polarisRegistry) delInstance(nodeID string) {
	delete(p.register, nodeID)
}

func (p *polarisRegistry) Options() registry.Options {
	return p.options
}
