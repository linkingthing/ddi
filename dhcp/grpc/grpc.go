package grpc

import (
	"context"
	"github.com/linkingthing/ddi/pb"
	"google.golang.org/grpc"
)

const (
	address = "localhost:8888"
)

func GetLeases(subNetID string) []*pb.Lease {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil
	}
	defer conn.Close()
	cli := pb.NewDhcpv4ManagerClient(conn)
	var target pb.GetLeasesReq
	target.Subnetid = subNetID
	var resp *pb.GetLeasesResp
	if resp, err = cli.GetLeases(context.Background(), &target); err != nil {
		return nil
	}
	return resp.Leases
}
