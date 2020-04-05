package GrpcConnectionPool

import (
	"log"
	"math/rand"
	"sync"
)

var poolOnce sync.Once

//连接池管理
type ConnectionPool struct {
	//读写锁,解决并发问题 暂不使用
	lock sync.RWMutex

	//串行化通道  存储有更新的健康服务列表
	queue chan *healthServiceList
	//key为服务名
	pool map[string]map[string]chan *grpcConnection

	//健康服务列表及其ip:port
	serviceMap map[string][]string
}

type healthServiceList struct {
	service     string
	addressList []string
}

func (this *ConnectionPool) newConnectionPool() {
	this.pool = make(map[string]map[string]chan *grpcConnection)
	this.serviceMap = make(map[string][]string)
	this.queue = make(chan *healthServiceList, SerializeQueueLength)
}

//初始化连接池和健康服务列表监控
func InitConnectionPool() *ConnectionPool {
	pool := &ConnectionPool{}
	poolOnce.Do(pool.newConnectionPool)
	go func() {
		pool.watch()
	}()
	return pool
}

func (this *ConnectionPool) createNewConnChan(address, service string) {
	log.Printf("发现新上线的微服务,服务名:%s,远程地址:%s,开始注册连接\n", service, address)
	c := make(chan *grpcConnection, ActiveConn)
	for i := 1; i <= ActiveConn; i++ {
		c <- NewGrpcConnection(address, service)
	}
	this.pool[service][address] = c
}

//随机负载均衡
func (this *ConnectionPool) PopConnection(service string) *grpcConnection {
	this.lock.RLock()
	defer this.lock.RUnlock()
	r := rand.Intn(len(this.serviceMap[service]))
	cc := <-this.pool[service][this.serviceMap[service][r]]
	log.Printf("从连接池获取连接,服务名:%s ,远程地址:%s\n:", service, cc.address)
	return cc
}

func (this *ConnectionPool) PutConnection(conn *grpcConnection) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.pool[conn.service][conn.address] <- conn
	log.Printf("归还连接,服务名:%s ,远程地址:%s\n:", conn.service, conn.address)
}

func (this *ConnectionPool) serviceListUpdate(list *healthServiceList) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if value, ok := this.pool[list.service]; ok {
		temp := make(map[string]struct{})
		for i := 0; i < len(list.addressList); i++ {
			temp[list.addressList[i]] = struct{}{}
			if _, have := value[list.addressList[i]]; !have {
				this.createNewConnChan(list.addressList[i], list.service)
			}
		}
		for k, _ := range value {
			if _, have := temp[k]; !have {
				this.closeConnChan(list.service, k)
			}
		}
	} else {
		this.pool[list.service] = make(map[string]chan *grpcConnection)
		for i := 0; i < len(list.addressList); i++ {
			this.createNewConnChan(list.addressList[i], list.service)
		}
	}
	this.serviceMap[list.service] = list.addressList
}

func (this *ConnectionPool) notice(list *healthServiceList) {
	this.queue <- list
}

func (this *ConnectionPool) watch() {
	for {
		health := <-this.queue
		this.serviceListUpdate(health)
	}
}

func (this *ConnectionPool) closeConnChan(service, address string) {
	log.Printf("发现关机的微服务,服务名:%s,远程地址:%s,关闭已注册连接\n", service, address)
	for v := range this.pool[service][address] {
		v.close()
	}
	delete(this.pool[service], address)
}
