package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

// NewUser 创建一个用户的API Create a user's API
func NewUser(conn net.Conn) *User {
	userAddr := conn.LocalAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
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
