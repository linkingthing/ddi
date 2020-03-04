package server

import (
	"encoding/json"
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

	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))
}
