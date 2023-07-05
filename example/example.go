package example

import (
	"discovery"
	"fmt"
	"log"
	"os"
	"os/signal"
)

type OrderService struct {
	name string
	add  string
}

var _ discovery.Service = &OrderService{}

func (o *OrderService) Name() string {
	return o.name
}

func (o *OrderService) Addr() string {
	return o.add
}

func OrderRegister(addr string) {
	endpoints := []string{"localhost:2379"}
	registrar, err := discovery.NewRegistrarEtcd(endpoints)
	if err != nil {
		log.Println("NewRegistrarEtcd err=", err)
		return
	}
	orderService := &OrderService{
		name: "order",
		add:  addr,
	}
	err = registrar.Register(orderService)
	if err != nil {
		log.Printf("register service err=%+v", err)
		return
	}
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, os.Interrupt)
	select {
	case <-closeChan:
		err = registrar.Deregister()
		if err != nil {
			fmt.Println("Deregister err=", err)
			return
		}
	}
}

func OrderDiscovery() {
	endpoints := []string{"localhost:2379"}
	discover, err := discovery.NewDiscoveryEtcd(endpoints)
	if err != nil {
		log.Println("NewDiscoveryEtcd err=", err)
		return
	}
	addr, err := discover.GetServiceAddr("order")
	if err != nil {
		log.Println("get service err=", err)
		return
	}
	fmt.Println("addr=", addr)
}
