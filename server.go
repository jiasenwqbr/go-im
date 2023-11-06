package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户列表  online user list
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	// 消息广播的channel   Channel for message broadcasting
	Message chan string
}

// NewServer 创建一个server的接口 Create an interface for a server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		// 将message发送给所有在线用户
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// BroadCast 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg

}
func (this *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	fmt.Println("连接建立成功")
	user := NewUser(conn, this)

	user.Online()
	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn read err:", err)
				return
			}
			// 提取用户消息（去除\n）
			msg := string(buf[:n-1])
			// 用户针对msg进行消息处理
			user.DoMessage(msg)
		}
	}()
	// 当前用户阻塞
	select {}

}

//启动服务器接口 Start Server Interface
func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net Listent err", err)
		return
	}
	// close listen port
	defer listener.Close()
	// 启动监听Message的goroutine
	go this.ListenMessager()
	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listener accept err:", err)
		}

		// do handler
		go this.Handler(conn)
	}
}
