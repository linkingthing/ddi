package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/linkingthing/ddi/pb"
	"google.golang.org/grpc"
)

var cmd = ""
var addr = ""

const (
	STARTDNS   = "StartDNS"
	STOPDNS    = "StopDNS"
	CREATEACL  = "CreateACL"
	DELETEACL  = "DeleteACL"
	CREATEVIEW = "CreateView"
	UPDATEVIEW = "UpdateView"
	DELETEVIEW = "DeleteView"
	CREATEZONE = "CreateZone"
	DELETEZONE = "DeleteZone"
	CREATERR   = "CreateRR"
	UPDATERR   = "UpdateRR"
	DELETERR   = "DeleteRR"
)

func init() {
	flag.StringVar(&cmd, "cmd", "", STARTDNS+"\n"+
		STOPDNS+"\n"+
		CREATEACL+"\n"+
		DELETEACL+"\n"+
		CREATEVIEW+"\n"+
		UPDATEVIEW+"\n"+
		DELETEVIEW+"\n"+
		CREATEZONE+"\n"+
		DELETEZONE+"\n"+
		CREATERR+"\n"+
		UPDATERR+"\n"+
		DELETERR)
	flag.StringVar(&addr, "addr", "localhost:8888", "ip:port")
}
func main() {
	flag.Parse()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	cli := pb.NewAgentManagerClient(conn)
	switch cmd {
	case STARTDNS:
		target := pb.DNSStartReq{Config: ""}
		_, err := cli.StartDNS(context.Background(), &target)
		if err != nil {
			fmt.Println(err)
		}
	case STOPDNS:
		target := pb.DNSStopReq{}
		OperResult, err := cli.StopDNS(context.Background(), &target)
		if err != nil {
			fmt.Println(OperResult)
			fmt.Println(err)
		} else {
			fmt.Println(OperResult)
		}
	}
}
