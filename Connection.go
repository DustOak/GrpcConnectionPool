package GrpcConnectionPool

import (
	"context"
	"google.golang.org/grpc"
	"log"
)

type grpcConnection struct {
	service    string
	address    string
	connection *grpc.ClientConn
}

//创建
func NewGrpcConnection(address, service string) *grpcConnection {
	ctx, _ := context.WithTimeout(context.TODO(), MaxWaitTime)
	//采用阻塞连接
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	return &grpcConnection{connection: conn, service: service, address: address}
}

func (this *grpcConnection) close() {
	this.connection.Close()
}

func (this *grpcConnection) GetClientConn() *grpc.ClientConn {
	return this.connection
}
