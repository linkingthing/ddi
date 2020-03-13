// refer to dns collector, dhcp statistics init here

package collector

import (
	"encoding/json"
	"github.com/linkingthing/ddi/utils"
	"log"
)

// unmarshall data from kea statistics commands
type CurlKeaArguments struct {
	Pkt4Received []interface{} `json:"pkt4-received"`
}
type CurlKeaStats struct {
	Arguments CurlKeaArguments `json:"arguments"`
	Result    string           `json:"result"`
}

type CurlKeaStatsAll struct {
	Arguments map[string]interface{} `json:"arguments"`
	Result    string                 `json:"result"`
}

// dashboard -- dhcp -- packet statistics
func (c *Metrics) GenerateDhcpPacketStatistics() error {
	//log.Println("+++ into GenerateDhcpPacketStatistics()")

	//get packet statistics data, export it to prometheus
	//todo move ip:port into conf
	url := "http://10.0.0.31:8000"
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`   {
	            "command": "statistic-get",
                "service": ["dhcp4"],
	            "arguments": {
	                "name": "pkt4-received"
	            }
	        }
	        ' 2>/dev/null`
	//log.Println("--- GenerateDhcpPacketStatistics curlCmd: ", curlCmd)
	out, err := utils.Cmd(curlCmd)

	if err != nil {
		log.Println("curl error: ", err)
		return err
	}
	//log.Println("+++ GenerateDhcpPacketStatistics(), out")
	//log.Println(out)
	//log.Println("--- GenerateDhcpPacketStatistics(), out")

	var curlRet CurlKeaStats
	json.Unmarshal([]byte(out[1:len(out)-1]), &curlRet)
	maps := curlRet.Arguments.Pkt4Received
	c.gaugeMetricData["dhcppacket"] = float64(len(maps))

	return nil
}

// dashboard -- dhcp -- packet statistics
func (c *Metrics) GenerateDhcpLeasesStatistics() error {
	//log.Println("+++ into GenerateDhcpPacketStatistics()")

	//get packet statistics data, export it to prometheus
	//todo move ip:port into conf
	url := "http://10.0.0.31:8000"
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`   {
                "command": "statistic-get-all",
                "service": ["dhcp4"],
                "arguments": { }
	        }
	        ' 2>/dev/null`
	//log.Println("--- GenerateDhcpPacketStatistics curlCmd: ", curlCmd)
	out, err := utils.Cmd(curlCmd)
	if err != nil {
		log.Println("curl error: ", err)
		return err
	}

	var curlRet CurlKeaStatsAll
	json.Unmarshal([]byte(out[1:len(out)-1]), &curlRet)

	maps := curlRet.Arguments
	for k, v := range maps {

		log.Println("in lease statistics(), for loop, k: ", k, ", v: ", v)
	}

	//maps := curlRet.Arguments.Pkt4Received
	c.gaugeMetricData["dhcplease"] = float64(33)

	return nil
}
