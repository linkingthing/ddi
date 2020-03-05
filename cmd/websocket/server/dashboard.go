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
type DashDns struct {
	Status  string    `json:"status"`
	Data    []Buckets `json:"data"`
	Message string    `json:"message"`
}

func NewDashDns() *DashDns {
	return &DashDns{}
}

//get statistics data from es
func GetDashDns(w http.ResponseWriter, r *http.Request) {

	//get node statistics data from es
	//默认显示最近1小时的统计数据
	r.ParseForm()
	log.Println("in Dash_DNS() Form: ", r.Form)

	EsServer := utils.EsServer + ":" + utils.EsPort + "/" + utils.EsIndex
	url := "http://" + EsServer + "/_search"
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
	log.Println("+++ GetDashDns(), out")
	log.Println(out)
	log.Println("--- GetDashDns(), out")
	if err != nil {
		log.Println("curl error: ", err)
		return
	}

	var result DashDns
	result.Status = "200"
	result.Message = "成功"

	var curlRetDash CurlRetDash

	//m := make(map[string]interface{})
	json.Unmarshal([]byte(out), &curlRetDash)

	//bytes := m["aggregations"].(map[string]interface{})["ips"].(map[string]interface{})["buckets"]
	log.Println("+++ print ips")
	log.Println(curlRetDash.Aggregations["ips"].Buckets)
	result.Data = curlRetDash.Aggregations["ips"].Buckets

	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))
}
