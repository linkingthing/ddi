package server

import (
	"encoding/json"
	"github.com/linkingthing/ddi/utils"
	"log"
	"net/http"
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
	log.Println("--- curlCmd: ", curlCmd)
	out, err := cmd(curlCmd)

	if err != nil {
		log.Println("curl error: ", err)
		return "", err
	}
	log.Println("+++ GetDashDnsIps(), out")
	log.Println(out)
	log.Println("--- GetDashDnsIps(), out")

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
	log.Println("--- GetDashDnsDomains curlCmd: ", curlCmd)
	out, err := cmd(curlCmd)

	if err != nil {
		log.Println("curl error: ", err)
		return "", err
	}
	log.Println("+++ GetDashDnsDomains(), out")
	log.Println(out)
	log.Println("--- GetDashDnsDomains(), out")

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
	log.Println("--- GetDashDnsResolveType curlCmd: ", curlCmd)
	out, err := cmd(curlCmd)

	if err != nil {
		log.Println("curl error: ", err)
		return "", err
	}
	log.Println("+++ GetDashDnsResolveType(), out")
	log.Println(out)
	log.Println("--- GetDashDnsResolveType(), out")

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
