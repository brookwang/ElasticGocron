package server

import (
	"errors"
	"fmt"
	"github.com/ouqiang/gocron/internal/modules/app"
	"github.com/ouqiang/gocron/internal/modules/rpc/auth"
	pb "github.com/ouqiang/gocron/internal/modules/rpc/proto"
	"github.com/ouqiang/gocron/internal/modules/sentinel"
	modules_service "github.com/ouqiang/gocron/internal/modules/service"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)



type NoticeServer struct{}

func (s NoticeServer) Run(ctx context.Context, req *pb.NoticeRequest) (*pb.NoticeResponse, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	log.Infof("execute cmd start: [cmd: %s]", req.Command)
	output, err := s.dealCommand(ctx, req.Command)
	resp := new(pb.NoticeResponse)
	resp.Output = output
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Error = ""
	}
	log.Infof("execute cmd end: [cmd: %s err: %s output:%s]", req.Command, resp.Error, resp.Output)

	return resp, nil
}

/*
执行处理命令
 */
func (s NoticeServer) dealCommand(ctx context.Context, command string) (string, error) {
	rpcPort := sentinel.GetRpcPort()
	port := app.Port
	//启动服务
	var shell string
	switch command {
	    case sentinel.StartDuplicate:
	   	  shell = app.AppDir + "/service.sh restart_gocron_duplicate " + port + " " + rpcPort
	    case sentinel.StartMaster:
	   	  shell = app.AppDir + "/service.sh duplicate_to_master " + port + " " + rpcPort
	    case sentinel.State:
          role := app.Role
          return role, nil
	   	default:
		  return "command error", nil
	}
	cmd := exec.Command("sh","-c", shell)
    //output, err := cmd.CombinedOutput()
    //用CombinedOutput杀死父进程后，子进程退出，用Run杀死父进程后，子进程仍存活，todo研究
    err := cmd.Run()
	if err != nil {
        return "", err
	}
	output := ""
	fmt.Println(cmd.Process.Pid)
    fmt.Println(string(output))
    return string(output), err
}

type UpdateConfigServer struct {

}

/**
修改配置文本
 */
func (s UpdateConfigServer) Run(ctx context.Context, req *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	log.Infof("execute cmd start: [ip: %s, command]", req.Ip, req.Command)
	if req.Ip == "" {
        return nil, errors.New("nil ip")
	}
	if req.Command == "" {
		return nil, errors.New("nil command")
	}
	var resUpdate bool
	switch req.Command {
	   case sentinel.UpdateMaster:
		   resUpdate = sentinel.UpdateSentinelConfig(req.Ip)
	   case sentinel.AddDuplicate:
		   resUpdate = modules_service.AddDuplicateConfig(req.Ip)
	default:
		return nil,errors.New("command error")
	}
	resp := new(pb.ConfigResponse)
	if resUpdate {
		//成功
		resp.Res = "0"
	}else {
		//失败
		resp.Res = "1"
	}
	log.Infof("execute cmd end: [command: %s  res:%s]", req.Command, resp.Res)
	return resp, nil
}




/*
哨兵
 */
func StartSentinel(addr string, enableTLS bool, certificate auth.Certificate) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepAliveParams),
		grpc.KeepaliveEnforcementPolicy(keepAlivePolicy),
	}
	if enableTLS {
		tlsConfig, err := certificate.GetTLSConfigForServer()
		if err != nil {
			log.Fatal(err)
		}
		opt := grpc.Creds(credentials.NewTLS(tlsConfig))
		opts = append(opts, opt)
	}
	server := grpc.NewServer(opts...)
	pb.RegisterNoticeServer(server, NoticeServer{})
	pb.RegisterUpdateConfigServer(server, UpdateConfigServer{})
	log.Infof("start sentinel server listen on %s", addr)

	go func() {
		err = server.Serve(l)
		if err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
		log.Infoln("收到信号 -- ", s)
		switch s {
		case syscall.SIGHUP:
			log.Infoln("收到终端断开信号, 忽略")
		case syscall.SIGINT, syscall.SIGTERM:
			log.Info("应用准备退出")
			server.GracefulStop()
			return
		}
	}

}


/*
开启哨兵服务的rpc服务
*/
func StartNoticeRpcServer(host string)  {
	var addr string
	port := sentinel.GetRpcPort()
	if host == "" {
		host = "0.0.0.0"
	}
	addr = host + ":" + port
	//启动grpc
	certificate := auth.Certificate{
		CAFile:   strings.TrimSpace(""),
		CertFile: strings.TrimSpace(""),
		KeyFile:  strings.TrimSpace(""),
	}
	StartSentinel(addr, false, certificate)
}
