package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// Consts
const (
	Message       = "Ping"
	StopCharacter = "\r\n\r\n"
)

var dddd string

// Start a server and process the request by a handler.
func startServer(port int, handler func(net.Conn)) {
	listen, err := net.Listen("tcp4", ":" + strconv.Itoa(port))
	defer listen.Close()
	if err != nil {
		log.Fatalf("Socket listen port %d failed,%s", port, err)
		os.Exit(1)
	}
	log.Printf("Begin listen port: %d", port)
	for {
		conn, err := listen.Accept() //监听端口，阻塞
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go handler(conn) //收到客户端连接请求，启动goroutine进行处理。
	}
}

// Handler of the leader node.
func leaderHandler(conn net.Conn) {
	defer conn.Close()
	var (
		buf = make([]byte, 1024)
		r   = bufio.NewReader(conn)
		w   = bufio.NewWriter(conn)
	)

	receivedMessage := ""
ILOOP:
	for {
		n, err := r.Read(buf)  //等待数据读取
		data := string(buf[:n])

		receivedMessage += data

		switch err {
		case io.EOF:
			break ILOOP
		case nil:
			log.Println("Leader Receive:", data) //打印读取的数据
			if isTransportOver(data) {
				break ILOOP
			}

		default:
			log.Fatalf("Receive data failed:%s", err)
			return
		}
	}
	receivedMessage = strings.TrimSpace(receivedMessage)
	ports := convertIntoInts(receivedMessage)
	ch := make(chan int)
	for i, port := range ports {
		go Send(port, strconv.Itoa(i), ch) //向多个客户端发送数据，使用chan收集客户端的响应数据。
	}
	count := 0
	for count < len(ports) { //等待所有响应数据并回显
		fmt.Printf("Slave %d has response the Ping message.\n", <-ch)
		count++
	}
	w.Write([]byte(Message))
	w.Flush()
	log.Printf("Leader Send: %d %s message.", count, Message)
}

// Helper library to convert '1,2,3,4' into []int{1,2,3,4}.
func convertIntoInts(data string) []int {
	var res = []int{}
	items := strings.Split(data, " ")
	for _, value := range items {
		intValue, err := strconv.Atoi(value)
		checkError(err)
		res = append(res, intValue)
	}
	return res
}

// Do check error.
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// SocketClient is to connect a socket given a port and send the given message.
func SocketClient(ip, message string, port int) (res string) {
	addr := strings.Join([]string{ip, strconv.Itoa(port)}, ":")
	conn, err := net.Dial("tcp", addr)

	defer conn.Close()

	if err != nil {
		log.Fatalln(err)
	}

	conn.Write([]byte(message))
	conn.Write([]byte(StopCharacter))
	log.Printf("Client Send: %s", Message)

	buff := make([]byte, 1024)
	n, _ := conn.Read(buff)
	res = string(buff[:n])
	return
}

// Send a message to another node with given port.
func Send(port int, message string, ch chan int) (returnMessage string) {
	ip := "127.0.0.1"
	returnMessage = SocketClient(ip, message, port)
	ch <- port
	fmt.Printf("Client response message: \"%s\"\n", returnMessage)
	return
}

// Handler for slave node.
func slaveHandler(conn net.Conn) {
	defer conn.Close()
	var (
		buf = make([]byte, 1024)
		r   = bufio.NewReader(conn)
		w   = bufio.NewWriter(conn)
	)

ILOOP:
	for {
		n, err := r.Read(buf)
		dddd = string(buf[:n])

		switch err {
		case io.EOF:
			break ILOOP
		case nil:
			log.Println("Slave Receive:", dddd)
			if isTransportOver(dddd) {
				break ILOOP
			}
		default:
			log.Fatalf("Receive data failed:%s", err)
			return
		}
	}
	res := fmt.Sprintf("I have received the message:[%s]", strings.TrimSpace(dddd))
	w.Write([]byte(res))
	w.Flush()
	//log.Printf("Slave Send: %s", Message)
}

func isTransportOver(data string) (over bool) {
	over = strings.HasSuffix(data, "\r\n\r\n")
	return
}

func main() {
	port := flag.Int("port", 3333, "port of the node.")
	mode := flag.String("mode", "leader", "should be slave or leader")
	flag.Parse()

	if strings.ToLower(*mode) == "leader" {
		// Start leader node.
		startServer(*port, leaderHandler)
	} else if strings.ToLower(*mode) == "slave" {
		// Start slave node.
		startServer(*port, slaveHandler)
	}
}
