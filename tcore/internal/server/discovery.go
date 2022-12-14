package tserver

import (
	"fmt"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/libkv/store"
	"github.com/rpcxio/rpcx-consul/client"
	"github.com/rpcxio/rpcx-consul/serverplugin"
	client2 "github.com/smallnest/rpcx/client"
	"reflect"
	"strings"
	"sync"
	tframework "tframework.com/rpc/tcore/interface"
	_interface "tframework.com/rpc/tcore/internal/interface"
	"tframework.com/rpc/tcore/internal/plugin"
	"time"
)

//***************************************************
//author tim.huang
//2022/8/10
//
//
//***************************************************

var ConsulDiscovery *TConsulServiceDiscovery
var initOne = &sync.Once{}

// TConsulServiceDiscovery
// @Description:
type TConsulServiceDiscovery struct {
	discovery     []*TmpDiscovery
	configService _interface.IServerConfigService
	lock          *sync.Mutex
}
type TmpDiscovery struct {
	moduleName string
	version    string
	discovery  *client.ConsulDiscovery
}

func (this *TConsulServiceDiscovery) GetDiscovery(moduleName, version string) *client.ConsulDiscovery {
	for _, discovery := range this.discovery {
		if discovery.moduleName == moduleName && discovery.version == version {
			return discovery.discovery
		}
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	//重复校验，避免并发场景下多次注册
	for _, discovery := range this.discovery {
		if discovery.moduleName == moduleName && discovery.version == version {
			return discovery.discovery
		}
	}
	//new discovery
	basePath := fmt.Sprintf("/tframework")
	servicePath := fmt.Sprintf("%v %v", moduleName, version)
	conf := &store.Config{
		ClientTLS:         nil,
		TLS:               nil,
		ConnectionTimeout: 0,
		Bucket:            "",
		PersistConnection: false,
		Username:          "",
		Password:          "",
	}
	d, _ := client.NewConsulDiscovery(basePath, servicePath, this.configService.GetConsulAddressSlice(), conf)
	data := &TmpDiscovery{
		moduleName: moduleName,
		version:    version,
		discovery:  d,
	}
	go func() {
		for {
			select {
			case kv := <-d.WatchService():
				for _, v := range kv {
					plugin.InfoS("consul watch service %v,%v", v.Key, v.Value)
				}
			}
		}
	}()
	this.discovery = append(this.discovery, data)
	return d
}

func (this *TConsulServiceDiscovery) RegisterServer(serviceAddress, moduleName string, module tframework.ITModule) (r *serverplugin.ConsulRegisterPlugin) {
	address := this.configService.GetConsulAddressSlice()
	r = &serverplugin.ConsulRegisterPlugin{
		ServiceAddress: "tcp@" + serviceAddress,
		ConsulServers:  address,
		BasePath:       this.configService.GetConsulPath(),
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Second * 11,
	}
	err := r.Start()
	if err != nil {
		plugin.InfoS("服务发现启动异常 %v", err)
	}
	return
}

func (this *TConsulServiceDiscovery) RegisterClient(service interface{}, moduleName, version string, cache map[string][]client2.XClient) {
	it := reflect.TypeOf(service)
	it = it.Elem()
	servicePath := fmt.Sprintf("%v %v", moduleName, version)
	discovery := this.GetDiscovery(moduleName, version)
	client := client2.NewXClient(servicePath, client2.Failover, client2.ConsistentHash, discovery, client2.DefaultOption)

	size := it.NumMethod()
	for i := 0; i < size; i++ {
		m := it.Method(i)
		if strings.HasPrefix(m.Name, rpcPrefix) {
			plugin.InfoS("注册 [%v:%v] 模块的 [%v] 接口", moduleName, version, m.Name)
			//proxyMethod := func(rpcType int32, args interface{}, reply interface{}) error {
			//	return client.Call(context.Background(), m.Name, &args, &reply)
			//}
			if cache[m.Name] == nil {
				cache[m.Name] = make([]client2.XClient, 0)
			}
			cache[m.Name] = append(cache[m.Name], client)
		}
	}
}

func instanceDefaultConsulDiscovery(configService _interface.IServerConfigService) {
	initOne.Do(
		func() {
			ConsulDiscovery = new(TConsulServiceDiscovery)
			ConsulDiscovery.discovery = make([]*TmpDiscovery, 0)
			ConsulDiscovery.configService = configService
			ConsulDiscovery.lock = new(sync.Mutex)
		})
}
