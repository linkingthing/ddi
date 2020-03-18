package main

import (
	"github.com/linkingthing/ddi/dhcp"
	"log"
	//"github.com/linkingthing/ddi/dhcp/server"
	"github.com/linkingthing/ddi/dhcp/server"
	"github.com/linkingthing/ddi/utils"
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
