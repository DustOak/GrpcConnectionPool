package GrpcConnectionPool

import "time"

var (
	//服务发现程序远程IP 默认本地
	IpAddress = "127.0.0.1"
	//服务发现程序端口 默认8500
	Port = 8500
	//保持的连接数
	ActiveConn = 30
	//连接等待时间
	MaxWaitTime = 5 * time.Second
	//串行化队列长度 当对订阅健康服务发现列表的对象发布新的信息时会进行串行化处理,该变量为串行化队列长度  默认10
	SerializeQueueLength = 10
)
