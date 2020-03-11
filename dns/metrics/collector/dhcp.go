// refer to dns collector, dhcp statistics init here

package collector

import (
	"log"
	"math/rand"
)

// dashboard -- dhcp -- packet statistics
func (c *Metrics) GenerateDhcpPacketStatistics() error {
	log.Println("+++ into GenerateDhcpPacketStatistics()")

	//get packet statistics data, export it to prometheus

	c.gaugeMetricData["dhcppacket"] = float64(rand.Int())

	return nil
}
