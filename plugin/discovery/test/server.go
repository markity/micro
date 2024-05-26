package test

import "fmt"

type MyPlugin struct {
}

func (*MyPlugin) Register(serviceName string, addrPort string) error {
	fmt.Println("注册")
	return nil
}

func (*MyPlugin) DeRegister() {
	fmt.Println("解除注册")
}
