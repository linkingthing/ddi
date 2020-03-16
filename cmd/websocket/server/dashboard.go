package server

import (
	"encoding/json"
	"fmt"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/dns/metrics/collector"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type Buckets struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}
type Ips struct {
	DocCount int `json:"doc_count_error_upper_bound"`
	SumOther int `json:"sum_other_doc_count"`
	Buckets  []Buckets
}
type CurlRetDash struct {
	Took         int            `json:"took"`
	TimedOut     bool           `json:"timed_out"`
	Hits         interface{}    `json:"hits"`
	Aggregations map[string]Ips `json:"aggregations"`
}
type DashDnsRet struct {
	Ips     []Buckets `json:"ips"`
	Domains []Buckets `json:"domains"`
	Types   []Buckets `json:"types"`
}
type DashDns struct {
	Status  string     `json:"status"`
	Data    DashDnsRet `json:"data"`
	Message string     `json:"message"`
}

func NewDashDns() *DashDns {
	return &DashDns{}
}

// dashboard - dhcp - subnet addresses assigned
type DhcpAssignStat struct {
	ID    json.Number `json:"id"`
	Name  string      `json:"name"`
	Addr  string      `json:"addr"`
	Total int         `json:"total"`
	Used  int         `json:"used"`
}

type BaseJsonDhcpAssign struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Data    []DhcpAssignStat `json:"data"`
}

func NewBaseJsonDhcpAssign() *BaseJsonDhcpAssign {
	return &BaseJsonDhcpAssign{}
}

func GetDashDnsIps(url string) (string, error) {
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`
{
    "size" : 0,
    "query" :{
    	"range": {
       	    "@timestamp" : 	{
       	    	"from": "now-37d"	
    	    }
    	}
    },
    "aggs" : {
        "ips" : { 
            "terms" : { 
              "field" : "ip"
            }
        }
    }
}
' 2>/dev/null 
`
	log.Println("in GetDashDnsIps(), curlCmd: ", curlCmd)

	out, err := utils.Cmd(curlCmd)
	if err != nil {
		log.Println("curl error: ", err)
		return "", err
	}

	return out, nil
}

func GetDashDnsDomains(url string) (string, error) {
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`
{
    "size" : 0,
    "query" :{
    	"range": {
       	    "@timestamp" : 	{
       	    	"from": "now-37d"	
    	    }
    	}
    },
    "aggs" : {
        "domains" : { 
            "terms" : { 
              "field" : "domain.keyword"
            }
        }
    }
}
' 2>/dev/null 
`

	out, err := utils.Cmd(curlCmd)

	if err != nil {
		log.Println("curl error: ", err)
		return "", err
	}

	return out, nil
}

// dns resolve types
func GetDashDnsResolveType(url string) (string, error) {
	curlCmd := "curl -X POST \"" + url + "\"" + " -H 'Content-Type: application/json' -d '" +
		`
{
    "size" : 0,
    "query" :{
    	"range": {
       	    "@timestamp" : 	{
       	    	"from": "now-37d"	
    	    }
    	}
    },
    "aggs" : {
        "types" : { 
            "terms" : { 
              "field" : "type.keyword"
            }
        }
    }
}
' 2>/dev/null 
`

	out, err := utils.Cmd(curlCmd)

	if err != nil {
		log.Println("curl error: ", err)
		return "", err
	}

	return out, nil
}

//get statistics data from es
func GetDashDns(w http.ResponseWriter, r *http.Request) {

	//get node statistics data from es
	log.Println("in Dash_DNS() Form: ", r.Form)

	EsServer := utils.EsServer + ":" + utils.EsPort + "/" + utils.EsIndex
	url := "http://" + EsServer + "/_search"

	var result DashDns
	result.Status = "200"
	result.Message = "成功"

	var curlRetDash CurlRetDash

	//get ip aggregations
	ips, err := GetDashDnsIps(url)
	if err != nil {
		log.Println("请求IP统计数据错误")
		return
	}
	json.Unmarshal([]byte(ips), &curlRetDash)
	log.Println("+++ print ips")
	log.Println(curlRetDash.Aggregations["ips"].Buckets)
	result.Data.Ips = curlRetDash.Aggregations["ips"].Buckets

	//get domain aggregations
	domains, err := GetDashDnsDomains(url)
	if err != nil {
		log.Println("域名统计数据错误")
		return
	}
	json.Unmarshal([]byte(domains), &curlRetDash)
	log.Println("+++ print domains")
	log.Println(curlRetDash.Aggregations["domains"].Buckets)
	result.Data.Domains = curlRetDash.Aggregations["domains"].Buckets

	//get resolve type aggregations
	types, err := GetDashDnsResolveType(url)
	if err != nil {
		log.Println("解析类型统计数据错误")
		return
	}
	json.Unmarshal([]byte(types), &curlRetDash)
	log.Println("+++ print resolve types")
	log.Println(curlRetDash.Aggregations["types"].Buckets)
	result.Data.Types = curlRetDash.Aggregations["types"].Buckets

	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))
}

func DashDhcpAssign(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	fmt.Println("in DashDhcpAssign(), Form: ", r.Form)

	//get servers maintained by kafka server
	result := NewBaseJsonDhcpAssign()

	result.Status = config.STATUS_SUCCCESS
	result.Message = config.MSG_OK

	assignMap := map[string]DhcpAssignStat{}
	// todo: get subnet total-address and assigned address from kea statistics api
	// define a temprary variable stores subnet id and total and used
	curlRet := collector.GetKeaStatisticsAll()
	maps := curlRet.Arguments
	stats := make(map[string]map[string]int)
	for k, v := range maps {
		//log.Println("in lease statistics(), for loop, k: ", k)
		rex := regexp.MustCompile(`^subnet\[(\d+)\]\.(\S+)`)
		out := rex.FindAllStringSubmatch(k, -1)

		if len(out) > 0 {
			var idx string
			var stat = DhcpAssignStat{}
			for _, i := range out {
				idx = i[1]
				addrType := i[2]
				//log.Println("get kea all, idx: ", idx, ", addrType: ", addrType)
				if stats[idx] == nil {
					stats[idx] = make(map[string]int)
				}

				if addrType == "total-addresses" {
					total := maps[k][0].([]interface{})[0]
					//log.Println("total: ", total)
					stats[idx]["total"] = int((collector.Decimal(total.(float64))))
				} else if addrType == "assigned-addresses" {
					stats[idx]["used"] = len(v)
				}
			}
			assignMap[idx] = stat
			log.Println("get kea all stats: ", stats)
		}
	}
	log.Println("stats: ", stats)

	//get subnet name and id from dhcp config
	k := dhcp.NewKEAv4Handler(dhcp.KEADHCPv4Service, dhcp.DhcpConfigPath, dhcp.Dhcpv4AgentAddr)
	conf := dhcp.ParseDhcpv4Config{}
	err := k.GetDhcpv4Config(dhcp.KEADHCPv4Service, &conf)
	if err != nil {
		log.Println("获取dhcp配置信息错误")
	}

	//log.Println("subnetv4 config: ", s4)
	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		//log.Println("k: ", k, ", v: ", v)

		var stat DhcpAssignStat
		stat.ID = v.Id
		stat.Name = v.Subnet + ":" + string(v.Id)
		stat.Addr = v.Subnet
		stat.Total = stats[string(v.Id)]["total"]
		stat.Used = stats[string(v.Id)]["used"]
		//stat.Total = assignMap[string(v.Id)].Total
		//stat.Used = assignMap[string(v.Id)].Used
		assignMap[string(v.Id)] = stat
		result.Data = append(result.Data, stat)
	}

	//log.Println("+++ result")
	//log.Println(result)
	//log.Println("--- result")

	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))

}
