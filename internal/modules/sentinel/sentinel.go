package sentinel

import (
	"errors"
	"fmt"
	"github.com/ouqiang/gocron/internal/modules/app"
	"github.com/ouqiang/gocron/internal/modules/logger"
	"github.com/ouqiang/gocron/internal/modules/net"
	"github.com/ouqiang/gocron/internal/modules/rpc/client"
	pb "github.com/ouqiang/gocron/internal/modules/rpc/proto"
	"github.com/ouqiang/gocron/internal/modules/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"path/filepath"
	"strings"
)

var sentinelConfig string  //哨兵配置文本路径
var NewMaster string   //选举出的新master
var RpcPort string //rpc端口

const (
	StopService = "stop"
	StartMaster = "start_master"
	StartDuplicate = "start_duplicate"
	State = "state"
)

const (
	UpdateMaster = "new_master"
	AddDuplicate = "add_duplicate"
)

/*
得到哨兵配置文本路径
 */
func GetSentinelConfigUri() string {
	return sentinelConfig
}

/*
检查主节点的状态
 */
func CheckMasterState() bool {
	//得到主节点IP
	ip, port := GetMasterInfo()
	masterState := net.CheckNodeState(ip, port)
	return masterState
}

/*
得到master的ip和端口
 */
func GetMasterInfo() (string, string) {
	sentinelConfigOb, _ := ini.Load(sentinelConfig)
	section := sentinelConfigOb.Section("node")
	masterInfo := section.Key("master").MustString("")
	arr := strings.Split(masterInfo, ":")
	Ip := arr[0]
	Port := arr[1]
	return Ip, Port
}

func SetSentinelConfigPath()  {
	sentinelConfig = filepath.Join(app.ConfDir, "/sentinel.ini")
	sentinelConfig = "/Users/baolin1/Documents/project/golang_module/conf/sentinel.ini"
}


/*
将某一副本转移为master
*/
func MoveMaster() bool {
	//选择切换身份的副本节点
    duplicateNode, err := SelectMasterFromDuplicate()
	if err != nil {
		logger.Error("select duolicateNode is nil")
		return false
	}
	NewMaster =  duplicateNode
	fmt.Println("选举出副本:" + duplicateNode + " 变更为主节点")
	//避免误判，通知主节点停止并以副本身份启动
    moveState := OldMasterMoveDuplicate()
	if moveState == false {
		logger.Error("OldMasterMoveDuplicate is false")
	}
	fmt.Println("主节点强制变更为副本节点")
    ip, port := utils.GetIPAndPortByAddress(duplicateNode)
    //将副本设为master节点
    setRes := SetNewMaster(ip, port)
	if setRes {
		logger.Info("SetNewMaster" + ip + ":"  + "success")
	}
	//确认转移结果
	moveState = ConfirmMoveState(ip, port)
	return moveState
}

/*
确认从节点转移为主节点的状态
return true是  fasle否
 */
func ConfirmMoveState(ip string,port string) bool {
	notice := new(pb.NoticeRequest)
	notice.Command = State
	res, err := client.NoticeExec(ip, port, notice)
	if err != nil {
		logger.Error("ConfirmMoveState fail "+ ip + ":" + port + ",info:" + err.Error())
		return false
	}
	//logger.Info有异常，待解决
	//logger.Info("ConfirmMoveState info" + ip + ":" + port + ",info:" + res)
	log.Infof("ConfirmMoveState info  %s : %s , info: %s ", ip, port, res)
	if res == "master" {
        return true
	}else {
		return false
	}
}


/*
从副本中选择主节点
 */
func SelectMasterFromDuplicate() (string, error) {
	//得到所有副本
	duplicateArr := getDuplicateList()
	var usableNode []string
	//检查存活的副本
	for _, value := range duplicateArr {
		valueArr := strings.Split(value, ":")
		ip := valueArr[0]
		port := valueArr[1]
		connRes := net.CheckNodeState(ip, port)
		if connRes {
			//可用的副本
            usableNode = append(usableNode, value)
		}
	}
	if len(usableNode) > 0 {
		//选择副本
		selectNode := usableNode[0]
		return selectNode, nil
	}else {
		err := errors.New("selectNode is nil")
		return "", err
	}
}

/*
得到副本列表
 */
func getDuplicateList() []string {
	sentinelConfigOb, _ := ini.Load(sentinelConfig)
	section := sentinelConfigOb.Section("node")
	duplicateInfo := section.Key("duplicate").MustString("")
	arr := strings.Split(duplicateInfo, ",")
	return arr
}

/*
得到rpc的端口
 */
func GetRpcPort() string {
	if RpcPort != "" {
        return RpcPort
	}
	sentinelConfigOb, _ := ini.Load(sentinelConfig)
	section := sentinelConfigOb.Section("rpc")
	port := section.Key("port").MustString("")
	return port
}

/*
得到默认的rpc端口
 */
func GetDefaultRpcPort() string {
	if sentinelConfig == "" {
		log.Fatal("sentinelConfig is nil")
	}
	fmt.Println(sentinelConfig)
	sentinelConfigOb, _ := ini.Load(sentinelConfig)
	section := sentinelConfigOb.Section("rpc")
	port := section.Key("port").MustString("")
	return port
}

/*
得到配置文件中的所有节点
 */
func GetAllNode() []string {
	sentinelConfigOb, _ := ini.Load(sentinelConfig)
	section := sentinelConfigOb.Section("node")
    master := section.Key("master").MustString("")
	duplicate := section.Key("duplicate").MustString("")
	masterArr := strings.Split(master, ",")
	duplicateArr := strings.Split(duplicate, ",")
	nodeArr := append(masterArr, duplicateArr...)
	return nodeArr
}



/*
停止master节点的服务，将master节点切换为副本
 */
func OldMasterMoveDuplicate() bool {
	return true
	//master状态正常则不执行
	state := CheckMasterState()
	if state {
		return false
	}
	ip, port := GetMasterInfo()
	notice := new(pb.NoticeRequest)
	notice.Command = StartDuplicate
	res, err := client.NoticeExec(ip, port, notice)
	if err != nil {
       logger.Error("OldMasterMoveDuplicate fail "+ ip + ":" + port + ",info:" + err.Error())
       return false
	}
	logger.Info("OldMasterMoveDuplicate info" + ip + ":" + port + ",info:" + res)
	return true
}

func UpateAllNodeConfig() bool {
	if NewMaster == "" {
		return false
	}
	res := UpdateSentinelConfig(NewMaster)
	if res == false {
        return false
	}
	go noticeAllNodeUpdateConfig()
	return true
}


/*
通知所有节点修改配置
 */
func noticeAllNodeUpdateConfig()  {
	nodeArr := GetAllNode()
	for _, address := range nodeArr {
		request := new(pb.ConfigRequest)
		request.Ip = NewMaster
		request.Command = UpdateMaster
		port := GetRpcPort()
		ip, _ := utils.GetIPAndPortByAddress(address)
		res, err := client.UpdateConfigExec(ip, port, request)
		if err != nil {
			logger.Error("NoticeAllNodeUpdateConfig fail "+ ip + ":" + port + ",err info:" + err.Error())
		}
		if res != "0" {
			logger.Error("NoticeAllNodeUpdateConfig fail "+ ip + ":" + port + ",res info:" + res)
		}else {
			logger.Info("NoticeAllNodeUpdateConfig success "+ ip + ":" + port + ",res info:" + res)
		}
	}
}


/*
更新集群各节点的配置文本
 */
func UpdateSentinelConfig(newmaster string) bool {
	if newmaster == "" {
		return false
	}

    //得到所有节点
	nodeArr := GetAllNode()
    masterArr := []string{newmaster}
	duplicateArr := utils.SliceDiff(nodeArr, masterArr)
	duplicateStr := strings.Join(duplicateArr,",")

	//读取当前配置文本内容
	sentinelConfigOb, _ := ini.Load(sentinelConfig)
	sectionRpc := sentinelConfigOb.Section("rpc")
	rpcConfig := []string{
		"port",sectionRpc.Key("port").MustString(""),
	}

	//新的node配置
	nodeConfig := []string{
		"master",newmaster,
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
	err = file.SaveTo(sentinelConfig)
	if err != nil {
        return false
	}
	return true
}




/*
设某一节点为master
 */
func SetNewMaster(ip string,port string) bool {
	notice := new(pb.NoticeRequest)
	notice.Command = StartMaster
	res, err := client.NoticeExec(ip, port, notice)
	if err != nil {
		logger.Error("SetNewMaster fail "+ ip + ":" + port + ",info:" + err.Error())
		return false
	}
	logger.Info("SetNewMaster info" + ip + ":" + port + ",info:" + res)
	return true
}




