# polaris

> Warning: Watch has not yet been implemented. If you want to use it please consider implementing it.

Polaris is a cloud-native service discovery and governance center. It can be used to solve the problem of service connection, fault tolerance, traffic control and secure in distributed and microservice architecture.

https://github.com/polarismesh/polaris

this plugin just use its registry and dicovery
in future maybe support more features,eg: Selector plugin for polaris's weight and priority

## usage
1. Do not support ListServices
2. in polaris_test, will do depoly this version: polaris-standalone-release_v1.11.3.linux.amd64.zip
3. according my test, it it necessary to call "time.Sleep(time.Second * x)" after call Register or Deregister, then the "GetService"  return correct data

```go
import sr "github.com/micro/plugins/v5/selector/registry"

func main() {
	poRegUrl := "127.0.0.1:8091"
	poRegNamespace := "default"
	poServiceToken := "nu/0WRA4EqSR1FagrjRj0fZwPXuGlMpX+zCuWu4uMqy8xr1vRjisSbA25aAC3mtU8MeeRsKhQiDAynUR09I="

	poReg := polaris.NewRegistry(
		registry.Addrs(poRegUrl),
			registry.Timeout(time.Second*5),
			// Can use Polaris's loadbalancer by Polairs web console
			polaris.GetOneInstance(true),
			polaris.NameSpace(poRegNamespace),
			polaris.ServerToken(poServiceToken),
		)

	// If polaris.GetOneInstance(true), its better to set sr.TTL(0)
	sel := sr.NewSelector(selector.Registry(poReg), sr.TTL(0))

	srv := micro.NewService()
	// Cmd will overwrite some option
	srv.Init()
	// Init twice
	srv.Init(
		micro.Registry(poReg),
		micro.Selector(sel),
		micro.RegisterInterval(time.Second*5),
		micro.RegisterTTL(time.Second*10),
	)
	// ...
}
```
