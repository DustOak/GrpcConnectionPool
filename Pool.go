package GrpcConnectionPool

import (
	"log"
	"math/rand"
	"sync"
)

var poolOnce sync.Once

//连接池管理
type ConnectionPool struct {
	////读写锁,解决并发问题 暂不使用
	//lock sync.RWMutex

	//key为服务名
	//pool map[string]map[string]chan *grpcConnection
	pool sync.Map

	//健康服务列表及其ip:port
	//serviceMap map[string][]string
	serviceMap sync.Map
}

type addressList struct {
	addressList []string
	len         int32
}

//func (this *ConnectionPool) newConnectionPool() {
//	//this.pool = make(map[string]map[string]chan *grpcConnection)
//	//this.serviceMap = make(map[string][]string)
//	trh
//}

//初始化连接池和健康服务列表监控
func InitConnectionPool() *ConnectionPool {
	pool := &ConnectionPool{}
	//poolOnce.Do(pool.newConnectionPool)
	return pool
}

func (this *ConnectionPool) createNewConnChan(address, service string) chan *grpcConnection {
	c := make(chan *grpcConnection, ActiveConn)
	for i := 1; i <= ActiveConn; i++ {
		c <- NewGrpcConnection(address, service)
	}
	return c
}

//随机负载均衡
func (this *ConnectionPool) PopConnection(service string) *grpcConnection {
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	addressMap, _ := this.serviceMap.Load(service)
	//if !ok {
	//	return nil, NoHaveService
	//}
	r := rand.Int31n(addressMap.(*addressList).len)
	services, _ := this.pool.Load(service)
	c, _ := (services.(*sync.Map)).Load(addressMap.(*addressList).addressList[r])
	log.Printf("从连接池获取连接,服务名:%s ,远程地址:%s\n:", service, c.(*grpcConnection).address)
	return c.(*grpcConnection)
}

func (this *ConnectionPool) PutConnection(conn *grpcConnection) {
	//this.lock.Lock()
	//defer this.lock.Unlock()
	addressList, _ := this.pool.Load(conn.service)
	c, _ := addressList.(*sync.Map).Load(conn.address)
	c.(chan *grpcConnection) <- conn
	log.Printf("归还连接,服务名:%s ,远程地址:%s\n:", conn.service, conn.address)

}

func (this *ConnectionPool) notice(service string, serviceAddressList []string) {
	if value, ok := this.pool.Load(service); ok {
		temp := make(map[string]struct{})
		for i := 0; i < len(serviceAddressList); i++ {
			temp[serviceAddressList[i]] = struct{}{}
			if _, ok := value.(*sync.Map).Load(serviceAddressList[i]); !ok {
				conn := this.createNewConnChan(serviceAddressList[i], service)
				value.(*sync.Map).Store(serviceAddressList[i], conn)
			}
		}
		value.(*sync.Map).Range(func(key, value interface{}) bool {
			_, ok := temp[key.(string)]
			if !ok {
				this.closeConnChan(service, key.(string))
			}
			return true
		})
	} else {
		var addressListMap *sync.Map
		for i := 0; i < len(serviceAddressList); i++ {
			conn := this.createNewConnChan(serviceAddressList[i], service)
			addressListMap.Store(serviceAddressList[i], conn)
		}
		this.pool.Store(service, addressListMap)
	}
	this.serviceMap.Store(service, &addressList{
		addressList: serviceAddressList,
		len:         int32(len(serviceAddressList)),
	})

	//this.lock.RLock()
	//if value, ok := this.pool[service]; ok {
	//	this.lock.RUnlock()
	//	temp := make(map[string]struct{})
	//	for i := 0; i < len(serviceAddressList); i++ {
	//		temp[serviceAddressList[i]] = struct{}{}
	//		this.lock.RLock()
	//		if _, ok := value[serviceAddressList[i]]; !ok {
	//			this.lock.RUnlock()
	//			conn := this.createNewConnChan(serviceAddressList[i], service)
	//			this.lock.Lock()
	//			value[serviceAddressList[i]] = conn
	//			this.lock.Unlock()
	//		}
	//	}
	//	//清理已经宕机的微服务的连接
	//	this.lock.RLock()
	//	for k, _ := range this.pool[service] {
	//		if _, have := temp[k]; !have {
	//			this.closeConnChan(service, k)
	//		}
	//	}
	//	this.lock.RUnlock()
	//} else {
	//	this.lock.RUnlock()
	//	this.lock.Lock()
	//	this.pool[service] = make(map[string]chan *grpcConnection)
	//	this.lock.Unlock()
	//	for i := 0; i < len(serviceAddressList); i++ {
	//		conn := this.createNewConnChan(serviceAddressList[i], service)
	//		this.lock.Lock()
	//		this.pool[service][serviceAddressList[i]] = conn
	//		this.lock.Unlock()
	//	}
	//}
	//this.lock.Lock()
	//this.serviceMap[service] = serviceAddressList
	//this.lock.Unlock()
}

func (this *ConnectionPool) closeConnChan(service, address string) {
	log.Println("有要关机的微服务,关闭已注册连接")
	value, _ := this.pool.Load(service)
	c, _ := value.(*sync.Map).Load(address)
	for v := range c.(chan *grpcConnection) {
		v.close()
	}
	value.(*sync.Map).Delete(address)
}
