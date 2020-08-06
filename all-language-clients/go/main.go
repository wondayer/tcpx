// Package go provides go client example
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/wondayer/tcpx"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	TCP_MESSAGE_ID_HANDSHAKE = 1 // 握手数据msgid
	TCP_MESSAGE_ID_NORMAL    = 2 // 业务数据msgid
	API_HOST_WEB             = "http://8.129.188.188:8000"
	TCP_SRV_HOST_WEB         = "8.129.188.188:8010"
	API_HOST_LOCAL           = "http://localhost:8000"
	TCP_SRV_HOST_LOCAL       = "localhost:8010"
)

func startTcpClient() {

	conn, e := net.Dial("tcp", TCP_SRV_HOST_LOCAL)
	if e != nil {
		panic(e)
	}

	sendbuf, e := tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: TCP_MESSAGE_ID_HANDSHAKE,
		Body:      []byte("a"),
	})
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	println(hex.Dump(sendbuf))

	_, e = conn.Write(sendbuf)
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	// 心跳间隔 10s
	var heartBeat []byte
	heartBeat, e = tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: tcpx.DEFAULT_HEARTBEAT_MESSAGEID,
		Body:      nil,
	})
	go func() {
		for {
			conn.Write(heartBeat)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		messageID, body, e := tcpx.UnPackFromReader(conn)
		if e != nil {
			fmt.Printf("unpack recvdata faild: %s", e.Error())
			return
		}
		switch messageID {
		case TCP_MESSAGE_ID_HANDSHAKE:
			log.Printf("receive handshake, message id: %d, body: %s\n", messageID, string(body))
			// 握手失败直接退出
			if strings.Contains(string(body), "success") {
				log.Print("tcp handshake success")
			} else {
				log.Print("tcp handshake faild, %s", body)
				os.Exit(-1)
			}
		case TCP_MESSAGE_ID_NORMAL:
			log.Printf("receive data, message id: %d, body: %s\n", messageID, string(body))
		case tcpx.DEFAULT_HEARTBEAT_MESSAGEID:
			log.Printf("receive heartbeat, message id: %d, body: %s\n", messageID, string(body))
		}
	}

	//scanner := bufio.NewScanner(conn) // reader为实现了io.Reader接口的对象，如net.Conn
	//scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	//	if !atEOF {
	//		var length = binary.BigEndian.Uint32(data[0:4]) // 数据部分长度
	//		if int(length)+4 <= len(data) { // 如果读取到的数据正文长度+4字节数据长度不超过读到的数据(实际上就是成功完整的解析出了一个包)
	//			return int(length) + 4, data[:int(length)+4], nil
	//		}
	//	}
	//	return
	//})
	//
	//// 打印接收到的数据包
	//for scanner.Scan() {
	//	DoMessage(scanner.Bytes())
	//}

}

func main() {

	startTcpClient()
}
