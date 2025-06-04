package consul

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
	hash "github.com/mitchellh/hashstructure"
	"go-micro.dev/v5/cmd"
	"go-micro.dev/v5/registry"
	mnet "go-micro.dev/v5/util/net"
)

type consulRegistry struct {
	Address []string
	opts    registry.Options

	client *consul.Client
	config *consul.Config

	// connect enabled
	connect bool

	queryOptions *consul.QueryOptions

	sync.Mutex
	register map[string]uint64
	// lastChecked tracks when a node was last checked as existing in Consul
	lastChecked map[string]time.Time
}

func init() {
	cmd.DefaultRegistries["consul"] = NewRegistry
}

func getDeregisterTTL(t time.Duration) time.Duration {
	// splay slightly for the watcher?
	splay := time.Second * 5
	deregTTL := t + splay

	// consul has a minimum timeout on deregistration of 1 minute.
	if t < time.Minute {
		deregTTL = time.Minute + splay
	}

	return deregTTL
}

func newTransport(config *tls.Config) *http.Transport {
	if config == nil {
		config = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     config,
	}
	runtime.SetFinalizer(&t, func(tr **http.Transport) {
		(*tr).CloseIdleConnections()
	})
	return t
}

func configure(c *consulRegistry, opts ...registry.Option) {
	// set opts
	for _, o := range opts {
		o(&c.opts)
	}

	// use default non pooled config
	config := consul.DefaultNonPooledConfig()

	if c.opts.Context != nil {
		// Use the consul config passed in the options, if available
		if co, ok := c.opts.Context.Value("consul_config").(*consul.Config); ok {
			config = co
		}
		if cn, ok := c.opts.Context.Value("consul_connect").(bool); ok {
			c.connect = cn
		}

		// Use the consul query options passed in the options, if available
		if qo, ok := c.opts.Context.Value("consul_query_options").(*consul.QueryOptions); ok && qo != nil {
			c.queryOptions = qo
		}
		if as, ok := c.opts.Context.Value("consul_allow_stale").(bool); ok {
			c.queryOptions.AllowStale = as
		}
	}

	// check if there are any addrs
	var addrs []string

	// iterate the options addresses
	for _, address := range c.opts.Addrs {
		// check we have a port
		addr, port, err := net.SplitHostPort(address)
		if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
			port = "8500"
			addr = address
			addrs = append(addrs, net.JoinHostPort(addr, port))
		} else if err == nil {
			addrs = append(addrs, net.JoinHostPort(addr, port))
		}
	}

	// set the addrs
	if len(addrs) > 0 {
		c.Address = addrs
		config.Address = c.Address[0]
	}

	if config.HttpClient == nil {
		config.HttpClient = new(http.Client)
	}

	// requires secure connection?
	if c.opts.Secure || c.opts.TLSConfig != nil {
		config.Scheme = "https"
		// We're going to support InsecureSkipVerify
		config.HttpClient.Transport = newTransport(c.opts.TLSConfig)
	}

	// set timeout
	if c.opts.Timeout > 0 {
		config.HttpClient.Timeout = c.opts.Timeout
	}

	// set the config
	c.config = config

	// remove client
	c.client = nil

	// setup the client
	c.Client()
}

func (c *consulRegistry) Init(opts ...registry.Option) error {
	configure(c, opts...)
	return nil
}

func (c *consulRegistry) Deregister(s *registry.Service, opts ...registry.DeregisterOption) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	// delete our hash and time check of the service
	c.Lock()
	delete(c.register, s.Name)
	delete(c.lastChecked, s.Name)
	c.Unlock()

	node := s.Nodes[0]

	return c.Client().Agent().ServiceDeregister(node.Id)
}

func (c *consulRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	var (
		regTCPCheck     bool
		regInterval     time.Duration
		regHTTPCheck    bool
		httpCheckConfig consul.AgentServiceCheck
		options         registry.RegisterOptions
	)
	for _, o := range opts {
		o(&options)
	}

	if c.opts.Context != nil {
		if tcpCheckInterval, ok := c.opts.Context.Value("consul_tcp_check").(time.Duration); ok {
			regTCPCheck = true
			regInterval = tcpCheckInterval
		}

		if conf, ok := c.opts.Context.Value("consul_http_check_config").(consul.AgentServiceCheck); ok {
			regHTTPCheck = true
			httpCheckConfig = conf
		}
	}

	// create hash of service; uint64
	h, err := hash.Hash(s, nil)
	if err != nil {
		return err
	}

	// use first node
	node := s.Nodes[0]

	// get existing hash and last checked time
	c.Lock()
	v, ok := c.register[s.Name]
	lastChecked := c.lastChecked[s.Name]
	c.Unlock()

	var check *consul.AgentServiceCheck
	checkTTL := regInterval
	if regTCPCheck {
		deregTTL := getDeregisterTTL(regInterval)

		check = &consul.AgentServiceCheck{
			TCP:                            node.Address,
			Interval:                       fmt.Sprintf("%v", regInterval),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%v", deregTTL),
		}

	} else if regHTTPCheck {
		interval, _ := time.ParseDuration(httpCheckConfig.Interval)
		deregTTL := getDeregisterTTL(interval)
		checkTTL = interval

		host, _, _ := net.SplitHostPort(node.Address)
		healthCheckURI := strings.Replace(httpCheckConfig.HTTP, "{host}", host, 1)

		check = &consul.AgentServiceCheck{
			HTTP:                           healthCheckURI,
			Interval:                       httpCheckConfig.Interval,
			Timeout:                        httpCheckConfig.Timeout,
			DeregisterCriticalServiceAfter: fmt.Sprintf("%v", deregTTL),
		}

		// if the TTL is greater than 0 create an associated check
	} else if options.TTL > time.Duration(0) {
		deregTTL := getDeregisterTTL(options.TTL)
		checkTTL = options.TTL

		check = &consul.AgentServiceCheck{
			TTL:                            fmt.Sprintf("%v", options.TTL),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%v", deregTTL),
		}
	}
	if c.opts.Context != nil {
		if ttl, ok := c.opts.Context.Value("consul_check_ttl").(time.Duration); ok {
			checkTTL = ttl
		}
	}

	// if it's already registered and matches then just pass the check
	if ok && v == h {
		passing := false
		if time.Since(lastChecked) > checkTTL {
			services, _, _ := c.Client().Health().Checks(s.Name, c.queryOptions)
			for _, service := range services {
				if service.ServiceID == node.Id && service.Status == "passing" {
					passing = true
					c.Lock()
					c.lastChecked[s.Name] = time.Now()
					c.Unlock()
					break
				}
			}
		} else {
			passing = true
		}
		if passing {
			if options.TTL == time.Duration(0) {
				return nil
			}
			// if the err is nil we're all good, bail out
			// if not, we don't know what the state is, so full re-register
			if err := c.Client().Agent().UpdateTTL("service:"+node.Id, "", "pass"); err == nil {
				return nil
			}
		}
		c.Deregister(s)
	}

	// encode the tags
	tags := encodeMetadata(node.Metadata)
	tags = append(tags, encodeEndpoints(s.Endpoints)...)
	tags = append(tags, encodeVersion(s.Version)...)

	host, pt, _ := net.SplitHostPort(node.Address)
	if host == "" {
		host = node.Address
	}
	port, _ := strconv.Atoi(pt)

	// register the service
	asr := &consul.AgentServiceRegistration{
		ID:      node.Id,
		Name:    s.Name,
		Tags:    tags,
		Port:    port,
		Address: host,
		Meta:    node.Metadata,
		Check:   check,
	}

	// Specify consul connect
	if c.connect {
		asr.Connect = &consul.AgentServiceConnect{
			Native: true,
		}
	}

	if err := c.Client().Agent().ServiceRegister(asr); err != nil {
		return err
	}

	// save our hash and time check of the service
	c.Lock()
	c.register[s.Name] = h
	c.lastChecked[s.Name] = time.Now()
	c.Unlock()

	// if the TTL is 0 we don't mess with the checks
	if options.TTL == time.Duration(0) {
		return nil
	}

	// pass the healthcheck
	return c.Client().Agent().UpdateTTL("service:"+node.Id, "", "pass")
}

func (c *consulRegistry) GetService(name string, opts ...registry.GetOption) ([]*registry.Service, error) {
	var rsp []*consul.ServiceEntry
	var err error

	// if we're connect enabled only get connect services
	if c.connect {
		rsp, _, err = c.Client().Health().Connect(name, "", false, c.queryOptions)
	} else {
		rsp, _, err = c.Client().Health().Service(name, "", false, c.queryOptions)
	}
	if err != nil {
		return nil, err
	}

	serviceMap := map[string]*registry.Service{}

	for _, s := range rsp {
		if s.Service.Service != name {
			continue
		}

		// version is now a tag
		version, _ := decodeVersion(s.Service.Tags)
		// service ID is now the node id
		id := s.Service.ID
		// key is always the version
		key := version

		// address is service address
		address := s.Service.Address

		// use node address
		if len(address) == 0 {
			address = s.Node.Address
		}

		svc, ok := serviceMap[key]
		if !ok {
			svc = &registry.Service{
				Endpoints: decodeEndpoints(s.Service.Tags),
				Name:      s.Service.Service,
				Version:   version,
			}
			serviceMap[key] = svc
		}

		var del bool

		for _, check := range s.Checks {
			// delete the node if the status is critical
			if check.Status == "critical" {
				del = true
				break
			}
		}

		// if delete then skip the node
		if del {
			continue
		}

		svc.Nodes = append(svc.Nodes, &registry.Node{
			Id:       id,
			Address:  mnet.HostPort(address, s.Service.Port),
			Metadata: decodeMetadata(s.Service.Tags),
		})
	}

	var services []*registry.Service
	for _, service := range serviceMap {
		services = append(services, service)
	}
	return services, nil
}

func (c *consulRegistry) ListServices(opts ...registry.ListOption) ([]*registry.Service, error) {
	rsp, _, err := c.Client().Catalog().Services(c.queryOptions)
	if err != nil {
		return nil, err
	}

	var services []*registry.Service

	for service := range rsp {
		services = append(services, &registry.Service{Name: service})
	}

	return services, nil
}

func (c *consulRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	return newConsulWatcher(c, opts...)
}

func (c *consulRegistry) String() string {
	return "consul"
}

func (c *consulRegistry) Options() registry.Options {
	return c.opts
}

func (c *consulRegistry) Client() *consul.Client {
	if c.client != nil {
		return c.client
	}

	for _, addr := range c.Address {
		// set the address
		c.config.Address = addr

		// create a new client
		tmpClient, _ := consul.NewClient(c.config)

		// test the client
		_, err := tmpClient.Agent().Host()
		if err != nil {
			continue
		}

		// set the client
		c.client = tmpClient
		return c.client
	}

	// set the default
	c.client, _ = consul.NewClient(c.config)

	// return the client
	return c.client
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	cr := &consulRegistry{
		opts:        registry.Options{},
		register:    make(map[string]uint64),
		lastChecked: make(map[string]time.Time),
		queryOptions: &consul.QueryOptions{
			AllowStale: true,
		},
	}
	configure(cr, opts...)
	return cr
}
