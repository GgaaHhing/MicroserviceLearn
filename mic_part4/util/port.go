package util

import (
	"net"
)

func GenRandomPort() int {
	//它的作用是将一个表示 TCP 网络地址的字符串解析为一个 *net.TCPAddr 类型的对象。
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	listen, err := net.Listen("tcp", addr.String())
	if err != nil {
		panic(err)
	}
	defer listen.Close()
	port := listen.Addr().(*net.TCPAddr).Port
	return port
}
