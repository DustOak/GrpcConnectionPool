package GrpcConnectionPool

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	"sync"
)

type Subscriber interface {
	notice(string, []string)
}

type Publisher interface {
	Subscript(subscriber Subscriber)
	publish(service string, serviceList []string)
}

type Watch interface {
	watchService(service string)
	watching()
}

var once sync.Once

type ServiceList struct {
	lock          sync.RWMutex
	subscriptList []Subscriber
	serviceList   []string
	config        *api.Config
	client        *api.Client
	address       string
}

//创建新的健康服务发布列表
func InitPublishServiceList(service []string) *ServiceList {
	address := fmt.Sprintf("%s:%d", IpAddress, Port)
	serviceList := &ServiceList{
		address:     address,
		serviceList: service,
	}
	once.Do(serviceList.newPublish)
	serviceList.watching()
	return serviceList
}

func (this *ServiceList) newPublish() {
	config := api.DefaultConfig()
	config.Address = this.address
	client, err := api.NewClient(config)
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	this.config = config
	this.client = client
	if err != nil {
		panic(err)
	}
	this.subscriptList = make([]Subscriber, 0)
}

//新增订阅
func (this *ServiceList) Subscript(subscriber Subscriber) {
	this.lock.Lock()
	this.subscriptList = append(this.subscriptList, subscriber)
	this.lock.Unlock()
}

//发布新健康列表
func (this *ServiceList) publish(service string, serviceList []string) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	for i := 0; i < len(this.subscriptList); i++ {
		this.subscriptList[i].notice(service, serviceList)
	}
}

func (this *ServiceList) watching() {
	for i := 0; i < len(this.serviceList); i++ {
		go this.watchService(this.serviceList[i])
	}
}

func (this *ServiceList) watchService(service string) {
	var lastIndex uint64
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	for {
		sv := make([]string, 0)
		services, metainfo, err := this.client.Health().Service(service, "", true, &api.QueryOptions{
			WaitIndex: lastIndex,
		})
		if err != nil {
			panic(err)
		}
		lastIndex = metainfo.LastIndex
		for i := 0; i < len(services); i++ {
			sv = append(sv, fmt.Sprintf("%s:%d", services[i].Service.Address, services[i].Service.Port))
		}
		this.publish(service, sv)
	}

}
