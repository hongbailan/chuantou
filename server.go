//server端，运行在有外网ip的服务器上
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
)

var localPort *string = flag.String("localPort", "3002", "user访问地址端口")
var remotePort *string = flag.String("remotePort", "20012", "与client通讯端口")

//与client相关的conn
type client struct {
	conn net.Conn
	er   chan bool
	//未收到心跳包通道
	heart chan bool
	//暂未使用！！！原功能tcp连接已经接通，不在需要心跳包
	disheart bool
	writ     chan bool
	recv     chan []byte
	send     chan []byte
}

//读取client过来的数据
func (self *client) read() {
	for {
		//40秒没有数据传输则断开
		self.conn.SetReadDeadline(time.Now().Add(time.Second * 40))
		var recv []byte = make([]byte, 10240)
		n, err := self.conn.Read(recv)

		if err != nil {
			//			if strings.Contains(err.Error(), "timeout") && self.disheart {
			//				fmt.Println("两个tcp已经连接,server不在主动断开")
			//				self.conn.SetReadDeadline(time.Time{})
			//				continue
			//			}
			self.heart <- true
			self.er <- true
			self.writ <- true
			//fmt.Println("长时间未传输信息，或者client已关闭。断开并继续accept新的tcp，", err)
		}
		//收到心跳包hh，原样返回回复
		if recv[0] == 'h' && recv[1] == 'h' {
			self.conn.Write([]byte("hh"))
			continue
		}
		self.recv <- recv[:n]

	}
}

//处理心跳包
//func (self client) cHeart() {

//	for {
//		var recv []byte = make([]byte, 2)
//		var chanrecv []byte = make(chan []byte)
//		self.conn.SetReadDeadline(time.Now().Add(time.Second * 30))
//		n, err := self.conn.Read(recv)
//		chanrecv <- recv
//		if err != nil {
//			self.heart <- true
//			fmt.Println("心跳包超时", err)
//			break
//		}
//		if recv[0] == 'h' && recv[1] == 'h' {
//			self.conn.Write([]byte("hh"))
//		}

//	}
//}

//把数据发送给client
func (self client) write() {

	for {
		var send []byte = make([]byte, 10240)
		select {
		case send = <-self.send:
			self.conn.Write(send)
		case <-self.writ:
			//fmt.Println("写入client进程关闭")
			break

		}
	}

}

//与user相关的conn
type user struct {
	conn net.Conn
	er   chan bool
	writ chan bool
	recv chan []byte
	send chan []byte
}

//读取user过来的数据
func (self user) read() {
	self.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 800))
	for {
		var recv []byte = make([]byte, 10240)
		n, err := self.conn.Read(recv)
		self.conn.SetReadDeadline(time.Time{})
		if err != nil {

			self.er <- true
			self.writ <- true
			//fmt.Println("读取user失败", err)

			break
		}
		self.recv <- recv[:n]
	}
}

//把数据发送给user
func (self user) write() {

	for {
		var send []byte = make([]byte, 10240)
		select {
		case send = <-self.send:
			self.conn.Write(send)
		case <-self.writ:
			//fmt.Println("写入user进程关闭")
			break

		}
	}

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	if flag.NFlag() != 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	local, _ := strconv.Atoi(*localPort)
	remote, _ := strconv.Atoi(*remotePort)
	if !(local >= 0 && local < 65536) {
		fmt.Println("端口设置错误")
		os.Exit(1)
	}
	if !(remote >= 0 && remote < 65536) {
		fmt.Println("端口设置错误")
		os.Exit(1)
	}

	//监听端口
	c, err := net.Listen("tcp", ":"+*remotePort)
	log(err)
	u, err := net.Listen("tcp", ":"+*localPort)
	log(err)
	//第一条tcp关闭或者与浏览器建立tcp都要返回重新监听
TOP:
	//监听user链接
	Uconn := make(chan net.Conn)
	go goaccept(u, Uconn)
	//一定要先接受client
	fmt.Println("准备好连接了")
	clientconnn := accept(c)
	fmt.Println("client已连接", clientconnn.LocalAddr().String())
	recv := make(chan []byte)
	send := make(chan []byte)
	heart := make(chan bool, 1)
	//1个位置是为了防止两个读取线程一个退出后另一个永远卡住
	er := make(chan bool, 1)
	writ := make(chan bool)
	client := &client{clientconnn, er, heart, false, writ, recv, send}
	go client.read()
	go client.write()

	//这里可能需要处理心跳
	for {
		select {
		case <-client.heart:
			goto TOP
		case userconnn := <-Uconn:
			//暂未使用
			client.disheart = true
			recv = make(chan []byte)
			send = make(chan []byte)
			//1个位置是为了防止两个读取线程一个退出后另一个永远卡住
			er = make(chan bool, 1)
			writ = make(chan bool)
			user := &user{userconnn, er, writ, recv, send}
			go user.read()
			go user.write()
			//当两个socket都创立后进入handle处理
			go handle(client, user)
			goto TOP
		}

	}

}

//监听端口函数
func accept(con net.Listener) net.Conn {
	CorU, err := con.Accept()
	logExit(err)
	return CorU
}

//在另一个进程监听端口函数
func goaccept(con net.Listener, Uconn chan net.Conn) {
	CorU, err := con.Accept()
	logExit(err)
	Uconn <- CorU
}

//显示错误
func log(err error) {
	if err != nil {
		fmt.Printf("出现错误： %v\n", err)
	}
}

//显示错误并退出
func logExit(err error) {
	if err != nil {
		//fmt.Printf("出现错误，退出线程： %v\n", err)
		runtime.Goexit()
	}
}

//显示错误并关闭链接，退出线程
func logClose(err error, conn net.Conn) {
	if err != nil {
		//fmt.Println("对方已关闭", err)
		runtime.Goexit()
	}
}

//两个socket衔接相关处理
func handle(client *client, user *user) {
	for {
		var clientrecv = make([]byte, 10240)
		var userrecv = make([]byte, 10240)
		select {

		case clientrecv = <-client.recv:
			user.send <- clientrecv
		case userrecv = <-user.recv:
			//fmt.Println("浏览器发来的消息", string(userrecv))
			client.send <- userrecv
			//user出现错误，关闭两端socket
		case <-user.er:
			//fmt.Println("user关闭了，关闭client与user")
			client.conn.Close()
			user.conn.Close()
			runtime.Goexit()
			//client出现错误，关闭两端socket
		case <-client.er:
			//fmt.Println("client关闭了，关闭client与user")
			user.conn.Close()
			client.conn.Close()
			runtime.Goexit()
		}
	}
}
