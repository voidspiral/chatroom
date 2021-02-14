package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":2021")
	if err != nil{
		panic(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleCon(conn)
	}

}

func handleCon(conn net.Conn) {
	defer conn.Close()

	//1.新用户进来，构建新用户实例
	user := &User{
		ID : GenUserID(),
		Addr: conn.RemoteAddr().String(),
		EnterAt:	time.Now(),
		MessageChannel : make(chan string, 8),
	}
	//2. 写
	go sendMessage(conn, user.MessageChannel)

	//3. 给当前用户发送欢迎信息，像所有用户告知新用户列表
	user.MessageChannel <- "Welcome, " + user.String()
	msg := Message{
		OwnerID : user.ID,
		Content : "user:`" + strconv.Itoa(user.ID) + "`has left",
	}
	messageChannel <- msg

	//4. 记录到全局用户列表，避免用锁
	enteringChannel <- user

	// 超时踢出
	var userActive = make(chan struct{})
	go func() {
		timer := time.NewTimer(time.Minute * 1)
		for {
			select {
			case <-timer.C:
				conn.Close()
			case <-userActive:
				timer.Reset(time.Minute * 1)
			}
		}
	}()
	//5.循环读取用户输入
	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg.Content = "user:`" + strconv.Itoa(user.ID) + "` has left"
		messageChannel <- msg

		// 用户活跃
		userActive <- struct{}{}
	}

	if err := input.Err(); err != nil {
		log.Println("读取错误: ", err)
	}
	//6. 用户离开
	leavingChannel <- user
	msg.Content = "user:`" + strconv.Itoa(user.ID) + "` has left"
	messageChannel <- msg

}

func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintf(conn, msg)
	}
}
// 生成用户ID
var (
	globalID int
	idLocker sync.Mutex
)

func GenUserID() int {
	idLocker.Lock()
	defer idLocker.Unlock()
	globalID++
	return globalID
}
type User struct {
	ID		int
	Addr	string
	EnterAt time.Time
	MessageChannel chan string
}
func (u *User) String() string {
	return u.Addr + ", UID:" + strconv.Itoa(u.ID) + ", Enter At:" +
		u.EnterAt.Format("2006-01-02 15:04:05+8000")
}
//给用户发送消息
type Message struct {
	OwnerID	int
	Content		string
}
var (
	//新用户到来, 通过channel进行登记
	enteringChannel = make(chan *User)
	// 用户离开，通过该 channel 进行登记
	leavingChannel = make(chan *User)
	// 广播专用的用户普通消息 channel，缓冲是尽可能避免出现异常情况堵塞
	messageChannel = make(chan Message, 8)
)
//broadcaster 用于记录聊天室用户，并进行消息广播
// 1. 新用户进来； 2.用户普通消息 3.用户消息
func broadcaster() {
	users := make(map[*User]struct{})
	for {
		select {
		case user := <-enteringChannel:
			users[user] = struct{}{}
		case user := <-leavingChannel:
			//用户离开
			delete(users, user)
			//避免 goroutine泄漏
			close(user.MessageChannel)
		case msg := <-messageChannel:
			for user := range users{
				if user.ID == msg.OwnerID {
					continue
				}
				user.MessageChannel <- msg.Content
			}
		}
	}
}