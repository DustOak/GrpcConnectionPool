package GrpcConnectionPool

import "errors"

var (
	NoHaveService           = errors.New("没有此服务")
	ServiceConnectionFailed = errors.New("微服务连接失败")
)
