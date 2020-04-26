package grpc

import (
	"context"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils"
	"google.golang.org/grpc"
)

func GetLeases(subNetID string) ([]*pb.Lease,error) {
	conn, err := grpc.Dial(utils.DHCPGrpcServer, grpc.WithInsecure())
	if err != nil {
		return nil,err
	}
	defer conn.Close()
	cli := pb.NewDhcpv4ManagerClient(conn)
	var target pb.GetLeasesReq
	target.Subnetid = subNetID
	var resp *pb.GetLeasesResp
	if resp, err = cli.GetLeases(context.Background(), &target); err != nil {
		return nil,err
	}
	return resp.Leases,nil
}
