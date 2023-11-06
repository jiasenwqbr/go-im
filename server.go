package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
	//监听用户是否活跃的channel
	isLive := make(chan bool)
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
			//用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}
	}()
	// 当前用户阻塞

	for {
		select {
		case <-isLive:
			fmt.Println("isLive")
			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 100):
			//已经超时
			//将当前的user强制关闭
			user.SendMsg("你被踢了")
			// 销毁用的资源
			close(user.C)
			// 关闭连接
			conn.Close()
			//退出当前的handler
			return
		}
	}

}

// Start 启动服务器接口 Start Server Interface
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
