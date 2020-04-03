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
)
