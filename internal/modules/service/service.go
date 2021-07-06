package service

import (
	"github.com/ouqiang/gocron/internal/modules/app"
	"github.com/ouqiang/gocron/internal/modules/logger"
	"github.com/ouqiang/gocron/internal/modules/rpc/client"
	pb "github.com/ouqiang/gocron/internal/modules/rpc/proto"
	"github.com/ouqiang/gocron/internal/modules/sentinel"
	"github.com/ouqiang/gocron/internal/modules/utils"
	"gopkg.in/ini.v1"
	"strings"
)

/*
服务注册(向主节点注册)
 */
func RegisterService(ip string, port int) bool {
	if !app.IsMasterRole() {
		return false
	}
	if ip == "0.0.0.0" {
		IP, err := utils.ExternalIP()
		if err != nil {
			return false
		}
		ip = string(IP)
	}
    address := ip + ":" + string(port)
	request := new(pb.ConfigRequest)
	request.Ip = address
	request.Command = sentinel.AddDuplicate
	masterIP, masterPort := sentinel.GetMasterInfo()
	res, err := client.UpdateConfigExec(masterIP, masterPort, request)
	if err != nil {
		logger.Error("RegisterService fail "+ ip + ":" + string(port) + ",err info:" + err.Error())
	}
	if res != "0" {
		logger.Error("RegisterService fail "+ ip + ":" + string(port) + ",res info:" + res)
	}else {
		logger.Info("RegisterService success "+ ip + ":" + string(port) + ",res info:" + res)
	}
	return true
}


/*
更新节点的配置文本
*/
func AddDuplicateConfig(ip string) bool {
	if ip == "" {
		return false
	}

	//读取当前配置文本内容
	sentinelConfigOb, _ := ini.Load(sentinel.GetSentinelConfigUri())

	//读取node内容
	section := sentinelConfigOb.Section("node")
	master := section.Key("master").MustString("")
	duplicate := section.Key("duplicate").MustString("")
	duplicateArr := strings.Split(duplicate, ",")

	//当前节点存在副本中
	if utils.InArray(duplicateArr, ip) {
		return true
	}

	duplicateArr = append(duplicateArr, ip)
	duplicateStr := strings.Join(duplicateArr,",")

    //读取rpc内容
	sectionRpc := sentinelConfigOb.Section("rpc")
	rpcConfig := []string{
		"port",sectionRpc.Key("port").MustString(""),
	}

	//新的node配置
	nodeConfig := []string{
		"master",master,
		"duplicate",duplicateStr,
	}

	//新的配置文本内容
	file := ini.Empty()
	sectionNodeNew, err := file.NewSection("node")
	if err != nil {
		return false
	}
	sectionRpcNew, err := file.NewSection("rpc")
	if err != nil {
		return false
	}
	for i := 0; i < len(nodeConfig); {
		_, err = sectionNodeNew.NewKey(nodeConfig[i], nodeConfig[i+1])
		if err != nil {
			return false
		}
		i += 2
	}
	for i := 0; i < len(rpcConfig); {
		_, err = sectionRpcNew.NewKey(rpcConfig[i], rpcConfig[i+1])
		if err != nil {
			return false
		}
		i += 2
	}
	err = file.SaveTo(sentinel.GetSentinelConfigUri())
	if err != nil {
		return false
	}
	return true
}
