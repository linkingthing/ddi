package main

import (
	"context"
	"os"
	"time"

	"log"

	"strconv"

	"fmt"

	"github.com/ben-han-cn/cement/shell"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/dhcp/server"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils"
	kg "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

func main() {
	//go Dhcpv4Client()

	utils.SetHostIPs() //set global vars from yaml conf

	//ver string, ConfPath string, addr string
	s, err := server.NewDHCPv4GRPCServer(dhcp.KEADHCPv4Service, dhcp.DhcpConfigPath, dhcp.Dhcpv4AgentAddr)
	if err != nil {

		log.Fatal(err)
		return
	}
	s.Start()
	defer s.Stop()

}
