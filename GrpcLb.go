package GrpcConnectionPool

import (
	"errors"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
	"net"
	"strconv"
)

//grpc 负载均衡库
func NewConsulResolver(address string, service string) naming.Resolver {
	return &consulResolver{
		address: address,
		service: service,
	}
}

type consulResolver struct {
	address string
	service string
}

func (r *consulResolver) Resolve(target string) (naming.Watcher, error) {
	config := api.DefaultNonPooledConfig()
	config.Address = r.address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &consulWatcher{
		client:  client,
		service: r.service,
		//addrs:   map[string]struct{}{},
	}, nil
}

type consulWatcher struct {
	client  *api.Client
	service string
	//addrs   map[string]struct{}
}

func (w *consulWatcher) Next() ([]*naming.Update, error) {
	services, _, err := w.client.Health().Service(w.service, "", true, &api.QueryOptions{})
	if err != nil {
		return nil, err
	}
	addrs := map[string]struct{}{}
	for _, service := range services {
		addrs[net.JoinHostPort(service.Service.Address, strconv.Itoa(service.Service.Port))] = struct{}{}
	}
	var updates []*naming.Update
	//for addr := range w.addrs {
	//	if _, ok := addrs[addr]; !ok {
	//		updates = append(updates, &naming.Update{Op: naming.Delete, Addr: addr})
	//	}
	//}
	for addr := range addrs {
		//if _, ok := w.addrs[addr]; !ok {
		updates = append(updates, &naming.Update{Op: naming.Add, Addr: addr})
		//}
	}
	if len(updates) != 0 {
		//w.addrs = addrs
		return updates, nil
	} else {
		return nil, errors.New("No Have" + w.service + "In Consul")
	}
}

func (w *consulWatcher) Close() {
}
