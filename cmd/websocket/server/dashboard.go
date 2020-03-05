package server

import (
	"encoding/json"
	"github.com/linkingthing/ddi/utils"
	"log"
	"net/http"
)

type DashDns struct {
	Status  string `json:"status"`
	Data    Nodes  `json:"data"`
	Message string `json:"message"`
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
	result := NewDashDns()

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
'
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

	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))
}
