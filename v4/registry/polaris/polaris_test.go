package polaris_test

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/go-micro/plugins/v4/registry/polaris"
	"go-micro.dev/v4/registry"
)

//go:generate echo abc

const deploy_polaris_scrips = `
#!bin/bash
version="1.11.3"
os="linux"
arch="amd64"
name="polaris-standalone-release_v$version.$os.$arch"
zipname="${name}.zip"
url="https://github.com/polarismesh/polaris/releases/download/v${version}/${zipname}" 
echo "$url"
# curl -o "./${zipname}" $url
if [ -f $zipname ];then
rm -rf $zipname
fi
if [ -d $name ];then
rm -rf $name
fi
wget $url
unzip ${zipname}
if [ ! -d $name ];then
echo "$name no exist"
exit 1
fi
cd ${name}
bash install.sh
ps -ef | grep -v grep |grep -E "polaris|prometheus"
ret=$(curl http://127.0.0.1:8090)
if [ "$ret" == "Polaris Server" ];then
echo "polaris start succ"
exit 0
else
echo "polaris start fail"
exit 1
fi
`

func init() {
	fmt.Println("polaris deploy")
	cmd := exec.Command("bash", "-c", deploy_polaris_scrips)
	bs, err := cmd.Output()
	fmt.Printf("polaris deploy err %v\n", err)
	fmt.Println(string(bs))
}

func genRegistry() registry.Registry {
	reg := polaris.NewRegistry(
		registry.Addrs("127.0.0.1:8091"),
		// registry.Addrs("172.17.72.159:8091"),
		registry.Timeout(time.Second*30),
		// Can use Polaris's loadbalancer by Polairs web console
		polaris.GetOneInstance(false),
		polaris.NameSpace("default"),
		polaris.ServerToken("nu/0WRA4EqSR1FagrjRj0fZwPXuGlMpX+zCuWu4uMqy8xr1vRjisSbA25aAC3mtU8MeeRsKhQiDAynUR09I="),
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
				Id:      name + version + addr,
				Address: addr,
			},
		},
	}
	return service
}

func TestRegister(t *testing.T) {
	//wait for Deregister
	//if Deregister be interrupt ,will effect other test with same address
	defer time.Sleep(time.Second * 10)
	//polairs reg wait time
	regWait := time.Second * 5
	reg := genRegistry()
	{
		srv := genService("test", "v1", "127.0.0.1:6000")
		assertNoError(t, reg.Register(&srv, registry.RegisterTTL(time.Second*30)))
		defer reg.Deregister(&srv)
		time.Sleep(regWait)
		services, err := reg.GetService("test")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
			t.SkipNow()
			return
		}

		for i, item := range services {
			t.Logf("--- 111 %d,%s,%v,%v,%v,%v\n", i, item.Name, item.Version, len(item.Nodes), item.Nodes[0].Address, item.Nodes[0].Id)
		}
		if len(services) != 1 {
			t.Errorf("expected %v, got %v", 1, len(services))
		}
	}

	{
		srv := genService("test", "v1", "127.0.0.1:7000")
		assertNoError(t, reg.Register(&srv, registry.RegisterTTL(time.Second*30)))
		defer reg.Deregister(&srv)
		time.Sleep(regWait)
		services, err := reg.GetService("test")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
			t.SkipNow()
			return
		}

		for i, item := range services {
			t.Logf("--- 222 %d,%s,%v,%v,%v,%v\n", i, item.Name, item.Version, len(item.Nodes), item.Nodes[0].Address, item.Nodes[0].Id)
		}
		if len(services) != 1 {
			t.Errorf("expected %v, got %v", 1, len(services))
		} else {
			if len(services[0].Nodes) != 2 {
				t.Errorf("expected %v, got %v", 2, len(services[0].Nodes))
			}
		}
	}

	{
		srv := genService("test", "v2", "127.0.0.1:8000")
		assertNoError(t, reg.Register(&srv, registry.RegisterTTL(time.Second*30)))
		defer reg.Deregister(&srv)
		time.Sleep(regWait)
		services, err := reg.GetService("test")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
			t.SkipNow()
			return
		}
		for i, item := range services {
			t.Logf("--- 333 %d,%s,%v,%v,%v,%v\n", i, item.Name, item.Version, len(item.Nodes), item.Nodes[0].Address, item.Nodes[0].Id)
		}
		if len(services) != 2 {
			t.Errorf("expected %v, got %v", 2, len(services))
		}
	}

}

func TestListService(t *testing.T) {
	//wait for Deregister
	//if Deregister be interrupt ,will effect other test with same address
	t.Logf("no support ListServices")

}

func TestDeregister(t *testing.T) {
	//wait for Deregister
	//if Deregister be interrupt ,will effect other test with same address
	defer time.Sleep(time.Second * 10)
	//polairs reg wait time
	regWait := time.Second * 5

	reg := genRegistry()
	service1 := genService("test-deregister", "v1", "127.0.0.1:6100")
	service2 := genService("test-deregister", "v2", "127.0.0.1:6101")

	assertNoError(t, reg.Register(&service1))
	time.Sleep(regWait)
	services, err := reg.GetService(service1.Name)
	assertNoError(t, err)
	assertEqual(t, 1, len(services))

	assertNoError(t, reg.Register(&service2))
	time.Sleep(regWait)
	services, err = reg.GetService(service2.Name)
	assertNoError(t, err)
	assertEqual(t, 2, len(services))

	assertNoError(t, reg.Deregister(&service1))
	time.Sleep(regWait)
	services, err = reg.GetService(service1.Name)
	assertNoError(t, err)
	assertEqual(t, 1, len(services))

	assertNoError(t, reg.Deregister(&service2))
	time.Sleep(regWait)
	services, err = reg.GetService(service1.Name)
	if err != registry.ErrNotFound {
		t.Error("expected err got nil")
	}
	assertEqual(t, 0, len(services))
}

func BenchmarkGetService(b *testing.B) {
	reg := genRegistry()
	srv := genService("one", "v1", "127.0.0.1:6200")
	assertNoError(b, reg.Register(&srv))
	defer reg.Deregister(&srv)
	time.Sleep(time.Second * 2)
	for n := 0; n < b.N; n++ {
		services, err := reg.GetService("one")
		assertNoError(b, err)
		assertEqual(b, 1, len(services))
		assertEqual(b, "one", services[0].Name)
	}
}
