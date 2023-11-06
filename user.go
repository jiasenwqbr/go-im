package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser 创建一个用户的API Create a user's API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.LocalAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动监听当前的user channel消息的goroutine  Start the goroutine that listens to the current user channel message
	go user.ListenMessage()
	return user
}

// ListenMessage 监听当前用户 channel的方法，一旦有消息，直接发送给客户端  The method of listening to the current user's channel, once there is a message, it is directly sent to the client
func (u User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}

// Online 用户上线
func (this *User) Online() {
	// 用户上线将用户加入到OnlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	// 广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

func (this *User) Offline() {
	// 用户上线将用户移除OnlineMap
	this.server.mapLock.Lock()
	//this.server.OnlineMap[this.Name] = this
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	// 广播当前用户下线消息
	this.server.BroadCast(this, "已下线")
}

func (this *User) DoMessage(msg string) {
	// 查询当前用户有哪些
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式: rename|张三
		newName := strings.Split(msg, "|")[1]
		// 判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名:" + this.Name + "\n")
		}
	} else {
		this.server.BroadCast(this, msg)
	}

}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}
