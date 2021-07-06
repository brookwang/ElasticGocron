package main

import (
	"fmt"
	"github.com/ouqiang/gocron/internal/modules/app"
	"github.com/ouqiang/gocron/internal/modules/sentinel"
	macaron "gopkg.in/macaron.v1"
	"time"
)

var AppVersion = "1.5"

/**
哨兵服务
 */
func main()  {
	//初始化配置文本
    initSentinelEnv()
    fmt.Println("start sentinel")
	var ch chan int
	//定时任务
	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for  {
			//从定时器中获取数据
			<-ticker.C
			deal()
		}
		ch <- 1
	}()
	<-ch
}

func deal()  {
	//检查主节点服务状态
	masterState := sentinel.CheckMasterState()
	//状态正常，跳过
	if masterState {
		fmt.Println("master node success")
		return
	}
	fmt.Println("master node fail")
	//出现客观失败，执行转移
	moveState := sentinel.MoveMaster()
	//未选举出新master
	if sentinel.NewMaster == "" {
		return
	}
	//未转移成功
	if !moveState {
		return
	}
	fmt.Println("转移从节点为主节点成功")
	//更新哨兵节点的配置
	updateRes := sentinel.UpateAllNodeConfig()
	if updateRes {
		fmt.Println("更新哨兵服务的配置文本为新主节点成功")
	}
}



/*
初始化哨兵环境
*/
func initSentinelEnv()  {
	app.InitEnv(AppVersion)
    macaron.Env = macaron.PROD
	sentinel.SetSentinelConfigPath()
}
