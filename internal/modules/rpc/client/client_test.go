package client

import (
	"fmt"
	pb "github.com/ouqiang/gocron/internal/modules/rpc/proto"
	"testing"
)

func TestNoticeExec(t *testing.T)  {
	//f, err := os.Create("trace_no.out")
	//defer f.Close()
	//err = trace.Start(f)
	noticeRequest := new(pb.NoticeRequest)
	noticeRequest.Command = "start_master"
	res, err := NoticeExec("10.222.76.101", "8888", noticeRequest)
	//defer trace.Stop()
	fmt.Println(err)
	fmt.Println(res)
}
