package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/linkingthing/ddi/pb"
	"google.golang.org/grpc"
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

	cli := pb.NewDhcpv6ManagerClient(conn)

	fmt.Printf("cmd: %s\n", cmd)
	switch cmd {

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
