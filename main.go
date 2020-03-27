package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"udpserver/config"
	"udpserver/controllers"
	_"udpserver/caches"
)

// 限制goroutine数量
var limitChan = make(chan bool, 5)

// UDP goroutine 实现并发读取UDP数据
func udpProcess(conn *net.UDPConn) {
	// 最大读取数据大小
	data := make([]byte, 1024)
	len, rAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Println("failed read udp msg, error: " + err.Error())
	} else {
		var req controllers.ReqJson
		err = json.Unmarshal(data[:len],&req)
		if err != nil {
			fmt.Println("Unmarshal error:",err)
		} else {
			rsp,err := controllers.Service(conn,rAddr,req)
			if err != nil {
				fmt.Println("Service error:",err)
			} else {
				rspBuff, err := json.Marshal(rsp)
				if err != nil {
					fmt.Println("Marshal error:",err)
				} else {
					_, err = conn.WriteToUDP(rspBuff, rAddr)
					if err != nil {
						fmt.Println("WriteToUDP error:",err)
					}
				}
			}
		}
	}
	<- limitChan
}

func udpServer(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println("read from connect failed, err:" + err.Error())
		os.Exit(1)
	}

	for {
		limitChan <- true
		go udpProcess(conn)
	}
}

func main() {
	addr := config.Conf.Get("common.Addr").(string)
	udpServer(addr)
}
