package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/pb"
	"google.golang.org/grpc"
)

var cmd string
var addr string

const (
	StartDHCPv4               = "StartDHCPv4"
	StopDHCPv4                = "StopDHCPv4"
	CreateSubnetv4            = "CreateSubnetv4"
	UpdateSubnetv4            = "UpdateSubnetv4"
	DeleteSubnetv4            = "DeleteSubnetv4"
	CreateSubnetv4Pool        = "CreateSubnetv4Pool"
	UpdateSubnetv4Pool        = "UpdateSubnetv4Pool"
	DeleteSubnetv4Pool        = "DeleteSubnetv4Pool"
	CreateSubnetv4Reservation = "CreateSubnetv4Reservation"
	UpdateSubnetv4Reservation = "UpdateSubnetv4Reservation"
	DeleteSubnetv4Reservation = "DeleteSubnetv4Reservation"
)

func init() {
	flag.StringVar(&cmd, "cmd", "", StartDHCPv4+"\n"+
		StopDHCPv4)
	flag.StringVar(&addr, "addr", dhcp.Dhcpv4AgentAddr, "ip:port")

}
func main() {
	flag.Parse()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Print(err)
		return
	}
	defer conn.Close()

	cli := pb.NewDhcpv4ManagerClient(conn)

	fmt.Printf("cmd: %s\n", cmd)
	switch cmd {
	case StartDHCPv4:
		target := pb.StartDHCPv4Req{Config: "StartDHCPv4Req"}
		_, err := cli.StartDHCPv4(context.Background(), &target)
		if err != nil {
			fmt.Print(err)
		} else {
			fmt.Print("start dhcpv4 ok\n")
		}
	case StopDHCPv4:
		target := pb.StopDHCPv4Req{}
		OperResult, err := cli.StopDHCPv4(context.Background(), &target)
		if err != nil {
			fmt.Print(OperResult)
			fmt.Print(err)
		} else {
			fmt.Print(OperResult)
		}

	}
}
