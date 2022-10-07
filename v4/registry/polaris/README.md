# polaris

Polaris is a cloud-native service discovery and governance center. It can be used to solve the problem of service connection, fault tolerance, traffic control and secure in distributed and microservice architecture.

https://github.com/polarismesh/polaris

this plugin just use its registry and dicovery
in future maybe support more features,eg: Selector plugin for polaris's weight and priority

## usage


```go
import sr "github.com/go-micro/plugins/v4/selector/registry"

poRegUrl := "127.0.0.1:8091"
poRegNamespace := "default"
poServiceToken := "nu/0WRA4EqSR1FagrjRj0fZwPXuGlMpX+zCuWu4uMqy8xr1vRjisSbA25aAC3mtU8MeeRsKhQiDAynUR09I="

poReg := polaris.NewRegistry(registry.Addrs(poRegUrl),
		registry.Timeout(time.Second*5),
		//can use polaris's loadbalance by polairs web console
		polaris.GetOneInstance(true),
		polaris.NameSpace(poRegNamespace),
		polaris.ServerToken(poServiceToken))

// if polaris.GetOneInstance(true) ,its better to set sr.TTL(0)
sel := sr.NewSelector(selector.Registry(poReg), sr.TTL(0))

srv := micro.NewService()
// cmd will overwrite some option
srv.Init()
// init twice
srv.Init(
	micro.Registry(poReg),
	micro.Selector(sel),
	micro.RegisterInterval(time.Second*5),
	micro.RegisterTTL(time.Second*10),
)

```
