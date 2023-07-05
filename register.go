package discovery

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

type Registrar interface {
	Register(service Service) error
	Deregister() error
}

type RegistrarEtcd struct {
	ctx      context.Context
	cli      *clientv3.Client
	leaseID  clientv3.LeaseID
	leaseTTL int64
}

var _ Registrar = &RegistrarEtcd{}

const (
	timeOut  = 3 * time.Second
	leaseTtl = 10
)

func NewRegistrarEtcd(endpoints []string) (*RegistrarEtcd, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("endpoints is nil")
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeOut,
	})
	if err != nil {
		return nil, err
	}
	re := &RegistrarEtcd{
		cli:      client,
		ctx:      context.Background(),
		leaseTTL: leaseTtl,
	}
	return re, nil
}

func (r *RegistrarEtcd) Register(service Service) error {
	// 1.申请租约
	grantRsp, err := r.cli.Grant(r.ctx, r.leaseTTL)
	if err != nil {
		return err
	}
	r.leaseID = grantRsp.ID
	// 2.写入etcd并绑定租约
	key := service.Name() + "_" + uuid.New().String()
	_, err = r.cli.Put(r.ctx, key, service.Addr(), clientv3.WithLease(r.leaseID))
	if err != nil {
		return err
	}
	// 3.续约
	aliveChan, err := r.cli.KeepAlive(r.ctx, r.leaseID)
	if err != nil {
		return err
	}
	go func() {
		for c := range aliveChan {
			log.Printf("keep alive, service name=%s,leaseID=%d\n", service.Name(), c.ID)
		}
	}()
	return nil
}

func (r *RegistrarEtcd) Deregister() error {
	// 1.撤销租约
	_, err := r.cli.Revoke(r.ctx, r.leaseID)
	if err != nil {
		return err
	}
	// 2. 关闭连接
	return r.cli.Close()
}
