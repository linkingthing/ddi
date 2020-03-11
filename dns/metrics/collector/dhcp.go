// refer to dns collector, dhcp statistics init here

package collector

import (
	"github.com/linkingthing/ddi/utils"
	"log"
)

// dashboard -- dhcp -- packet statistics
func (c *Metrics) GenerateDhcpPacketStatistics() error {
	log.Println("+++ into GenerateDhcpPacketStatistics()")

	//get packet statistics data, export it to prometheus
	url := "http://10.0.0.31:8000"
	//    curl -X POST "http://10.0.0.31:8000" -H 'Content-Type: application/json' -d '
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`   {
	            "command": "statistic-get",
                "service": ["dhcp4"],
	            "arguments": {
	                "name": "pkt4-received"
	            }
	        }
	        ' 2>/dev/null`
	log.Println("--- GenerateDhcpPacketStatistics curlCmd: ", curlCmd)
	out, err := utils.Cmd(curlCmd)

	if err != nil {
		log.Println("curl error: ", err)
		return err
	}
	log.Println("+++ GenerateDhcpPacketStatistics(), out")
	log.Println(out)
	log.Println("--- GenerateDhcpPacketStatistics(), out")

	c.gaugeMetricData["dhcppacket"] = float64(2.33)

	return nil
}
