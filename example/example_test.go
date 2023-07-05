package example

import "testing"

func TestRegistrar(t *testing.T) {
	go OrderRegister(":8080")
	go OrderRegister(":8081")
	go OrderRegister(":8082")
	select {}
}

func TestOrderDiscovery(t *testing.T) {
	for i := 0; i < 10; i++ {
		OrderDiscovery()
	}
}
