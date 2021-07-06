#!/bin/bash


APP_NAME=./bin/gocron
Port=$2
RpcPort=$3

usage() {
    echo "Usage: sh service.sh [start_gocron_master|stop_gocron|gocron_status|restart_gocron_duplicate|restart_gocron_master|build_gocron]"
    exit 1
}

#生成环境编译
build_gocron(){
   echo "start build"
   go build  -o ${APP_NAME}  cmd/gocron/gocron.go
   result=$?
   if [ $result != 0 ]; then
       echo "build error"
       exit $result
   fi
   echo "end build"
}




#停止方法
stop_duplicate_gocron(){
  check_params
  is_duplicate_gocron_exist
  if [ $? -eq "0" ]; then
    kill -9 $pid
  else
    echo "${APP_NAME} is not running"
  fi
}

#停止方法
stop_master_gocron(){
  check_params
  is_master_gocron_exist
  if [ $? -eq "0" ]; then
    kill -9 $pid
  else
    echo "${APP_NAME} is not running"
  fi
}



check_params(){
  if [${Port} == ""]; then
      echo "Port is nil"
      #强制结束本脚本
      kill -9 $$
  fi
  if [${RpcPort} == ""]; then
      echo "RpcPort is nil"
      #强制结束本脚本
      kill -9 $$
  fi
}



#以master身份启动Gocron服务
start_gocron_master(){
  check_params
  is_master_gocron_exist
  if [ $? -eq 0 ]; then
    echo "${APP_NAME} is already running. pid=${pid}"
  else
    echo "${APP_NAME} web  -r master -port ${Port} -sentinel_port ${RpcPort} &"
    nohup ${APP_NAME} web  -r master -port ${Port} -sentinel_port ${RpcPort} &
  fi
}

#以duplicate身份启动Gocron服务
start_gocron_duplicate(){
  check_params
  is_gocron_exist
  if [ $? -eq 0 ]; then
    echo "${APP_NAME} is already running. pid=${pid}"
  else
    echo "${APP_NAME} web -port ${Port} -sentinel_port ${RpcPort} &"
    nohup ${APP_NAME} web -port ${Port} -sentinel_port ${RpcPort} &
  fi
}

#以duplicate身份重启gocron
restart_gocron_duplicate(){
  check_params
  stop_duplicate_gocron
  sleep 1
  start_gocron_duplicate
}

#检查master程序是否在运行
is_master_gocron_exist(){
  check_params
  pid=`ps -ef|grep $APP_NAME.*web.*-r.*master.*-port.*${Port}.*-sentinel_port.*${RpcPort}|grep -v grep|awk '{print $2}'`
  #如果不存在返回1，存在返回0
  if [ -z "${pid}" ]; then
    return 1
  else
    return 0
  fi
}



#检查程序是否在运行
is_duplicate_gocron_exist(){
  check_params
  echo "$APP_NAME.*web.*-port.*${Port}.*-sentinel_port.*${RpcPort}"
  pid=`ps -ef|grep $APP_NAME.*web.*-port.*${Port}.*-sentinel_port.*${RpcPort}|grep -v grep|awk '{print $2}'`
  #如果不存在返回1，存在返回0
  if [ -z "${pid}" ]; then
    return 1
  else
    return 0
  fi
}


#输出运行状态
gocron_duplicate_status(){
  check_params
  is_duplicate_gocron_exist
  if [ $? -eq "0" ]; then
    echo "${APP_NAME} is running. Pid is ${pid}"
  else
    echo "${APP_NAME} is NOT running."
  fi
}


#以master身份重启gocron
restart_gocron_master(){
  check_params
  stop_master_gocron
  sleep 1
  start_gocron_master
}

duplicate_to_master(){
  check_params
  stop_duplicate_gocron
  sleep 1
  start_gocron_master
}




#根据输入参数，选择执行对应方法，不输入则执行使用说明
case "$1" in
  "start_gocron_master")
    start_gocron_master
    ;;
  "stop_master_gocron")
    stop_master_gocron
    ;;
  "gocron_duplicate_status")
    gocron_duplicate_status
    ;;
  "restart_gocron_master")
    restart_gocron_master
    ;;
  "restart_gocron_duplicate")
    restart_gocron_duplicate
    ;;
  "duplicate_to_master")
    duplicate_to_master
    ;;
  "build_gocron")
    build_gocron
    ;;
  *)
    usage
    ;;
esac





