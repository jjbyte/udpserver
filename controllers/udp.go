package controllers

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"udpserver/caches"
)

type ClientInfo struct {
	Addr		string `json:"addr"`
	ClientName	string `json:"name"`
}

type ReqJson struct {
	Command 	string `json:"command"`
	OwnerName	string `json:"owner"`
	EndName		string `json:"end"`
}

type RspJson struct {
	Result 		bool `json:"result"`
	Msg			interface{} `json:"msg"`
}

func parseAddr(addr string) net.UDPAddr {
	t := strings.Split(addr, ":")
	port, _ := strconv.Atoi(t[1])
	return net.UDPAddr{
		IP:   net.ParseIP(t[0]),
		Port: port,
	}
}

func Service (conn *net.UDPConn,addr *net.UDPAddr,req ReqJson) (rsp *RspJson,err error) {

	rsp = &RspJson{
		Result:false,
		Msg:"命令码错误",
	}

	//注册
	if req.Command == "1" {
		ipInfo := addr.String()
		val,err := caches.Cache.HSet("rd_client_info",req.OwnerName, ipInfo).Result()
		if err != nil {
			rsp.Result = false
			rsp.Msg = err.Error()
		}

		if val  {
			//注册成功
			rsp.Result = true
			rsp.Msg = "注册成功"
		} else {
			rsp.Result = false
			rsp.Msg = "注册失败"
		}
	}

	//获取所有用户列表
	if req.Command == "2" {
		clientInfo,err := caches.Cache.HGetAll("rd_client_info").Result()
		if err != nil {
			rsp.Result = false
			rsp.Msg = err.Error()
		} else {
			rsp.Result = true
			var infoList [] * ClientInfo
			for key,value := range clientInfo {
				info := new(ClientInfo)
				info.ClientName = key
				info.Addr = value
				infoList = append(infoList, info)
			}
			rsp.Msg = infoList
		}
	}

	//client a - > b
	if req.Command == "3" {
		//找到B的地址和端口
		clientInfo,err := caches.Cache.HGet("rd_client_info",req.EndName).Result()
		if err != nil {
			rsp.Result = false
			rsp.Msg = err.Error()
		} else {
			//告诉B，说A想和它通讯
			rsp.Result = true
			info := &ClientInfo{
				Addr:addr.String(),
				ClientName:req.OwnerName,
			}
			rsp.Msg = info
			rspBuff, err := json.Marshal(rsp)
			if err != nil {
				fmt.Println(err)
				rsp.Result = false
				rsp.Msg = "通知失败"
			} else {
				bAddr := parseAddr(clientInfo)
				_, err = conn.WriteToUDP(rspBuff, &bAddr)
				if err != nil {
					fmt.Println(err)
					rsp.Result = false
					rsp.Msg = "通知失败"
				} else {
					//将B的信息返回
					rsp.Result = true
					info := &ClientInfo{
						Addr:clientInfo,
						ClientName:req.EndName,
					}
					rsp.Msg = info
				}
			}
		}
	}

	return rsp,nil
}
