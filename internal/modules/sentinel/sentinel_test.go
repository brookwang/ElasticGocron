package sentinel

import (
	"fmt"
	"strconv"
	"testing"
)

func TestGetRpcPort(t *testing.T)  {
	SetSentinelConfigPath()
	res := GetRpcPort()
	fmt.Println(res)
}

func TestConfirmMoveState(t *testing.T)  {
	SetSentinelConfigPath()
	res := ConfirmMoveState("127.0.0.1", "5821")
	fmt.Println(res)
}

func TestGetAllNode(t *testing.T)  {
	SetSentinelConfigPath()
	res := GetAllNode()
	fmt.Println(res)
}

func TestUpdateSentinelConfig(t *testing.T) {
	SetSentinelConfigPath()
	res := UpdateSentinelConfig("127.0.0.3")
	fmt.Println(res)
}


func TestString(t *testing.T)  {
	var port int
	port = 80
	str := strconv.Itoa(port)
	fmt.Println(str)
}


