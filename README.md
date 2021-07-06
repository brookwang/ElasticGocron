# 项目背景
本项目fork于项目https://github.com/ouqiang/gocron。

# 二次开发的目的
* gocron的调度器仅支持单机部署，有单点风险。本项目将调度器从单点改造为分布式部署，调度器分为master主节点和副本duplicate节点。
* 主节点负责任务调度，副本节点仅提供备份。
* 单独的哨兵服务监控主节点服务状态，出现异常情况，下线主节点，选举副本节点为主节点负责任务调度，实现调度器的高可用。


# gocron - 定时任务管理系统
[![Downloads](https://img.shields.io/github/downloads/ouqiang/gocron/total.svg)](https://github.com/ouqiang/gocron/releases)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)](https://github.com/ouqiang/gocron/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/ouqiang/gocron.svg?label=Release)](https://github.com/ouqiang/gocron/releases)

# 项目简介
使用Go语言开发的轻量级定时任务集中调度和管理系统, 用于替代Linux-crontab [查看文档](https://github.com/ouqiang/gocron/wiki)

原有的延时任务拆分为独立项目[延迟队列](https://github.com/ouqiang/delay-queue)  

## 功能特性
* Web界面管理定时任务
* crontab时间表达式, 精确到秒
* 任务执行失败可重试
* 任务执行超时, 强制结束
* 任务依赖配置, A任务完成后再执行B任务
* 账户权限控制
* 任务类型
    * shell任务
    > 在任务节点上执行shell命令, 支持任务同时在多个节点上运行
    * HTTP任务
    > 访问指定的URL地址, 由调度器直接执行, 不依赖任务节点
* 查看任务执行结果日志
* 任务执行结果通知, 支持邮件、Slack、Webhook

### 截图
![流程图](https://raw.githubusercontent.com/ouqiang/gocron/master/assets/screenshot/scheduler.png)
![任务](https://raw.githubusercontent.com/ouqiang/gocron/master/assets/screenshot/task.png)
![Slack](https://raw.githubusercontent.com/ouqiang/gocron/master/assets/screenshot/notification.png)
    
### 支持平台
> Windows、Linux、Mac OS

### 环境要求
>  MySQL


## 下载

## 安装

###  二进制安装
1. 解压压缩包   
2. `cd 解压目录`   
3. 启动        
* 调度器启动        
  * Windows: `gocron.exe web`   
  * Linux、Mac OS:  `./gocron web`
* 任务节点启动, 默认监听0.0.0.0:5921
  * Windows:  `gocron-node.exe`
  * Linux、Mac OS:  `./gocron-node`
4. 浏览器访问 http://localhost:5920

### 源码安装

- 安装Go 1.11+
- `go get -d github.com/brookwang/ElasticGocron`
- `export GO111MODULE=on` 
- 编译 `make`
- 启动
    * 调度器主节点 gocron `./bin/gocron web --role master`
    * 调度器副本节点 gocron `./bin/gocron web`
    * 执行器 gocron-node `./bin/gocron-node`
  


### 开发

1. 安装Go1.9+, Node.js, Yarn
2. 安装前端依赖 `make install-vue`
3. 启动gocron, gocron-node `make run`
4. 启动node server `make run-vue`, 访问地址 http://localhost:8080

访问http://localhost:8080, API请求会转发给gocron

`make` 编译

`make run` 编译并运行

`make package` 打包 
> 生成当前系统的压缩包 gocron-v1.5-darwin-amd64.tar.gz gocron-node-v1.5-darwin-amd64.tar.gz

`make package-all` 生成Windows、Linux、Mac的压缩包

### 命令

* gocron
    * -v 查看版本

* gocron web
    * --host 默认0.0.0.0
    * --role 指定调度器角色, master|duplicate, master为主节点, duplicate为副本节点，默认duplicate  
    * -p 端口, 指定端口, 默认5920
    * -e 指定运行环境, dev|test|prod, dev模式下可查看更多日志信息, 默认prod
    * -h 查看帮助
  
* gocron-node
    * -allow-root *nix平台允许以root用户运行
    * -s ip:port 监听地址  
    * -enable-tls 开启TLS    
    * -ca-file   CA证书文件   
    * -cert-file 证书文件  
    * -key-file  私钥文件
    * -h 查看帮助
    * -v 查看版本

## To Do List
- [x] 版本升级
- [x] 批量开启、关闭、删除任务
- [x] 调度器与任务节点通信支持https
- [x] 任务分组
- [x] 多用户
- [x] 权限控制

## 程序使用的组件
* Web框架 [Macaron](http://go-macaron.com/)
* 定时任务调度 [Cron](https://github.com/robfig/cron)
* ORM [Xorm](https://github.com/go-xorm/xorm)
* UI框架 [Element UI](https://github.com/ElemeFE/element)
* 依赖管理 [Govendor](https://github.com/kardianos/govendor)
* RPC框架 [gRPC](https://github.com/grpc/grpc)

## 反馈

## ChangeLog


