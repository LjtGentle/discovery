package discovery

import (
	"context"
	"errors"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"log"
	"math/rand"
)

type Discovery interface {
	GetServiceAddr(serviceName string) (string, error)
	WatchService(serviceName string) error
}

type DiscoveryEtcd struct {
	cli *clientv3.Client
	ctx context.Context
}

func NewDiscoveryEtcd(endpoints []string) (*DiscoveryEtcd, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeOut,
	})
	if err != nil {
		return nil, err
	}
	d := &DiscoveryEtcd{
		cli: client,
		ctx: context.Background(),
	}
	return d, nil
}

var _ Discovery = &DiscoveryEtcd{}

func (d *DiscoveryEtcd) GetServiceAddr(serviceName string) (string, error) {
	// 前缀获取
	getRes, err := d.cli.Get(d.ctx, serviceName, clientv3.WithPrefix())
	if err != nil {
		return "", err
	}
	if len(getRes.Kvs) == 0 {
		return "", errors.New(fmt.Sprintf("service %s not found", serviceName))
	}
	// 随机轮训
	randInt := rand.Intn(len(getRes.Kvs))
	log.Println("randInt=", randInt)
	addr := string(getRes.Kvs[randInt].Value)

	return addr, nil
}

func (d *DiscoveryEtcd) WatchService(serviceName string) error {
	watchChan := d.cli.Watch(d.ctx, serviceName, clientv3.WithPrefix())
	go func() {
		for watch := range watchChan {
			log.Printf("watch=%+v", watch)
		}
	}()
	return nil
}
