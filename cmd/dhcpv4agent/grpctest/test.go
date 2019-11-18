package main

import (
	"context"
	"flag"
	"fmt"

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
	StartDHCPv6               = "StartDHCPv6"
	StopDHCPv6                = "StopDHCPv6"
)

func init() {
	flag.StringVar(&cmd, "cmd", "", StartDHCPv4+"\n"+
		StopDHCPv4+"\n"+
		StartDHCPv6+"\n"+
		StopDHCPv6)
	flag.StringVar(&addr, "addr", "localhost:8888", "ip:port")

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
		target := pb.StartDHCPv4Req{Config: ""}
		_, err := cli.StartDHCPv4(context.Background(), &target)
		if err != nil {
			fmt.Println(err)
		}
	case StopDHCPv4:
		target := pb.StopDHCPv4Req{}
		OperResult, err := cli.StopDHCPv4(context.Background(), &target)
		if err != nil {
			fmt.Println(OperResult)
			fmt.Println(err)
		} else {
			fmt.Println(OperResult)
		}

	case StartDHCPv6:
		fmt.Print("\n---into case startdhcpv6---\n")
		target := pb.StartDHCPv6Req{Config: ""}
		_, err := cli.StartDHCPv6(context.Background(), &target)
		if err != nil {
			fmt.Println(err)
		}
	case StopDHCPv6:
		fmt.Print("\n---into case StopDHCPv6---\n")
		target := pb.StopDHCPv6Req{}
		OperResult, err := cli.StopDHCPv6(context.Background(), &target)
		if err != nil {
			fmt.Println(OperResult)
			fmt.Println(err)
		} else {
			fmt.Println(OperResult)
		}
	}
}
