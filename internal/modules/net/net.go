package net

import (
	"net"
	"time"
)

/*
检查节点状态
检查存活返回 true, 失败false
*/
func CheckNodeState(ip string, port string) bool {
	address := ip + ":" + port
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return false
	}
	if conn != nil {
		conn.Close()
		return true
	} else {
		return false
	}
}
