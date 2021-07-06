package client

import (
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"sync"
	"time"

	"google.golang.org/grpc/status"

	"github.com/ouqiang/gocron/internal/modules/logger"
	"github.com/ouqiang/gocron/internal/modules/rpc/grpcpool"
	pb "github.com/ouqiang/gocron/internal/modules/rpc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

var (
	taskMap sync.Map
)

const backOffMaxDelay = 3 * time.Second


var (
	errUnavailable = errors.New("无法连接远程服务器")
)

var (
	keepAliveParams = keepalive.ClientParameters{
		Time:                20 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}
)

func generateTaskUniqueKey(ip string, port int, id int64) string {
	return fmt.Sprintf("%s:%d:%d", ip, port, id)
}

func Stop(ip string, port int, id int64) {
	key := generateTaskUniqueKey(ip, port, id)
	cancel, ok := taskMap.Load(key)
	if !ok {
		return
	}
	cancel.(context.CancelFunc)()
}

func Exec(ip string, port int, taskReq *pb.TaskRequest) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic#rpc/client.go:Exec#", err)
		}
	}()
	addr := fmt.Sprintf("%s:%d", ip, port)
	c, err := grpcpool.Pool.Get(addr)
	if err != nil {
		return "", err
	}
	if taskReq.Timeout <= 0 || taskReq.Timeout > 86400 {
		taskReq.Timeout = 86400
	}
	timeout := time.Duration(taskReq.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	taskUniqueKey := generateTaskUniqueKey(ip, port, taskReq.Id)
	taskMap.Store(taskUniqueKey, cancel)
	defer taskMap.Delete(taskUniqueKey)

	resp, err := c.Run(ctx, taskReq)
	if err != nil {
		return parseGRPCError(err)
	}

	if resp.Error == "" {
		return resp.Output, nil
	}

	return resp.Output, errors.New(resp.Error)
}


/*
哨兵服务rpc通知
 */
func NoticeExec(ip string, port string, noticeRequest *pb.NoticeRequest) (string, error) {
	address := ip + ":" + port
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepAliveParams),
		grpc.WithBackoffMaxDelay(backOffMaxDelay),
		grpc.WithInsecure(),
		//该参数保证连接重置
		grpc.WithBlock(),
	}
	timeout := time.Duration(3) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return "", err
	}
	defer conn.Close()
    c := pb.NewNoticeClient(conn)
    r, err := c.Run(context.Background(), noticeRequest)
	if err != nil {
		return "", err
	}
	return r.Output, nil
}


/*
哨兵服务的修改配置rpc通知
*/
func UpdateConfigExec(ip string, port string, configRequest *pb.ConfigRequest) (string, error) {
	address := ip + ":" + port
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepAliveParams),
		grpc.WithBackoffMaxDelay(backOffMaxDelay),
		grpc.WithInsecure(),
	}
	timeout := time.Duration(3) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	c := pb.NewUpdateConfigClient(conn)
	r, err := c.Run(context.Background(), configRequest)
	if err != nil {
		return "", err
	}
	return r.Res, nil
}


func parseGRPCError(err error) (string, error) {
	switch status.Code(err) {
	case codes.Unavailable:
		return "", errUnavailable
	case codes.DeadlineExceeded:
		return "", errors.New("执行超时, 强制结束")
	case codes.Canceled:
		return "", errors.New("手动停止")
	}
	return "", err
}
