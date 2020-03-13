package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type Metric struct {
	Name     string `json:"__name__"`
	Instance string
	Job      string
}

type ValueIntf [2]interface {
}
type ValueIntfOne interface {
}

type Result struct {
	Metric Metric
	Value  []ValueIntfOne
	Values []ValueIntf
}
type Data struct {
	ResultType string
	Result     []Result
}
type Response struct {
	Status string
	Data   Data
}

// v2
type MyHandler struct{}

type Usage struct {
	Cpu  string `json:"cpu"`
	Mem  string `json:"mem"`
	Disk string `json:"disk"`
	Qps  string `json:"qps"`
}

type NodeType struct {
	nodeType map[string]utils.PromRole
}

type Hosts struct {
	Nodes map[string]utils.PromRole `json:"nodes"`
	Usage map[string]Usage          `json:"usage"`
}

type BaseJsonBean struct {
	Status  string `json:"status"`
	Data    Hosts  `json:"data"`
	Message string `json:"message"`
}

//type values interface {
//}

type RangeMetric struct {
	Node string `json:"node"`
}

type Nodes struct {
	Metric RangeMetric `json:"metric"`
	Values interface{} `json:"values"`
}

type BaseJsonRange struct {
	Status  string `json:"status"`
	Data    Nodes  `json:"data"`
	Message string `json:"message"`
}

// node management module, list servers
type BaseJsonServer struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Data    []utils.PromRole `json:"data"`
}

func NewBaseJsonBean() *BaseJsonBean {
	return &BaseJsonBean{}
}
func NewBaseJsonRange() *BaseJsonRange {
	return &BaseJsonRange{}
}
func NewBaseJsonServer() *BaseJsonServer {
	return &BaseJsonServer{}
}

func (*MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}

func Query_range(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("in query_range() Form: ", r.Form)
	result := NewBaseJsonRange()

	paramStart, _ := r.Form["start"]
	paramEnd, _ := r.Form["end"]
	paramStep, _ := r.Form["step"]
	paramHost, _ := r.Form["node"]
	paramType, _ := r.Form["type"]

	if paramStart == nil || paramEnd == nil || paramStep == nil || paramHost == nil || paramType == nil {
		fmt.Println("ERROR, param need to be checked")
		result.Status = "error"
		result.Message = "params not sent"
		bytes, _ := json.Marshal(result)
		//fmt.Fprint(w, string(bytes))
		w.Write([]byte(bytes))
		return
	}

	start, err := strconv.Atoi(paramStart[0])
	end, err := strconv.Atoi(paramEnd[0])
	step, err := strconv.Atoi(paramStep[0])
	host := paramHost[0]
	t := paramType[0]

	result.Status = "success"
	result.Message = "ok"
	result.Data.Metric.Node = host

	//
	//cpuResp, err := GetPromRange("cpu", "10.0.0.55", 1579150980, 1579154580, 323)
	cpuResp, err := GetPromRange(t, host, start, end, step)
	if err != nil {
		log.Println(err)
		return
	}
	var histData []interface{}
	err = json.Unmarshal([]byte(*cpuResp), &histData)
	if err != nil {
		log.Println("cpuResp unmarshal error ", err)
	}

	result.Data.Values = histData

	log.Println("xxx cpuHist: ", cpuResp)

	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))
}

func Query(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	fmt.Println("Form: ", r.Form)
	//for k, v := range r.Form {
	//	fmt.Println(k, "=>", v, strings.Join(v, "-"))
	//}
	host, _ := r.Form["node"]
	promType, _ := r.Form["type"]

	log.Println("host: ", host)
	log.Println("promType: ", promType)

	result := NewBaseJsonBean()

	result.Status = config.STATUS_SUCCCESS
	result.Message = config.MSG_OK
	result.Data.Nodes = utils.OnlinePromHosts

	cpuResp, err := GetPromItem("cpu", utils.PromLocalInstance)
	if err != nil {
		log.Println(err)
		var hosts Hosts
		result.Status = config.STATUS_ERROR
		result.Message = config.ERROR_PROM_CPU
		result.Data = hosts

		bytes, _ := json.Marshal(result)
		//fmt.Fprint(w, string(bytes))
		w.Write([]byte(bytes))
		return
	}

	cpuUsage, err := strconv.ParseFloat(cpuResp, 64)
	cpuResp = fmt.Sprintf("%.2f", cpuUsage)
	log.Println("cpuResp: ", cpuResp)

	memResp, err := GetPromItem("mem", utils.PromLocalInstance)
	if err != nil {
		log.Println(err)
		var hosts Hosts
		result.Status = config.STATUS_ERROR
		result.Message = config.ERROR_PROM_MEM
		result.Data = hosts

		bytes, _ := json.Marshal(result)
		//fmt.Fprint(w, string(bytes))
		w.Write([]byte(bytes))
		return
	}
	log.Println("memResp: ", memResp)
	memUsage, err := strconv.ParseFloat(memResp, 64)
	memResp = fmt.Sprintf("%.2f", memUsage)

	diskResp, err := GetPromItem("disk", utils.PromLocalInstance)
	if err != nil {
		log.Println(err)
		var hosts Hosts
		result.Status = config.STATUS_ERROR
		result.Message = config.ERROR_PROM_DISK
		result.Data = hosts

		bytes, _ := json.Marshal(result)
		//fmt.Fprint(w, string(bytes))
		w.Write([]byte(bytes))
		return
	}
	log.Println("diskResp: ", diskResp)
	diskUsage, err := strconv.ParseFloat(diskResp, 64)
	diskResp = fmt.Sprintf("%.2f", diskUsage)

	var HostUsage = make(map[string]Usage)
	var Usage = Usage{}
	Usage.Cpu = cpuResp
	Usage.Mem = memResp
	Usage.Disk = diskResp
	Usage.Qps = strconv.Itoa(rand.Intn(99)) + "." + strconv.Itoa(rand.Intn(99))
	postHost := ""
	if host != nil {
		postHost = host[0]
	}

	log.Println("postHost: ", postHost)
	if postHost != "" {

		HostUsage[postHost] = Usage
	} else {
		result.Status = config.STATUS_ERROR
		result.Message = config.ERROR_PARAM_HOST
		bytes, _ := json.Marshal(result)
		//fmt.Fprint(w, string(bytes))
		w.Write([]byte(bytes))
		return
	}
	log.Println("hostUsage: ", HostUsage)
	result.Data.Usage = HostUsage

	//log.Println("+++ result")
	//log.Println(result)
	//log.Println("--- result")
	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))

	return
}

// getProm get prometheus data from prometheus api
func GetPromItem(promType string, host string) (string, error) {
	var command string
	var rsp Response
	var promWebHost = utils.PromServer + ":" + utils.PromPort
	var url = "http://" + promWebHost + "/api/v1/query?query="
	if promType == "cpu" {
		command = "curl -H \"Content-Type: application/json\" http://" + promWebHost +
			"/api/v1/query?query=instance:node_cpu:avg_rate5m 2>/dev/null"
	} else if promType == "disk" {
		promStr := "100%20-%20(node_filesystem_free_bytes{mountpoint=\"/\",fstype=~\"ext4|xfs\"}%20/%20node_filesystem_size_bytes{mountpoint=\"/\",fstype=~\"ext4|xfs\"}%20*%20100)"
		command = "curl -g '" + url + promStr + "' 2>/dev/null"
	} else if promType == "mem" {
		promStr := "(node_memory_MemFree_bytes%2Bnode_memory_Cached_bytes%2Bnode_memory_Buffers_bytes)%20/%20node_memory_MemTotal_bytes%20*%20100"
		command = "curl -g '" + url + promStr + "' 2>/dev/null"
	}
	log.Println("in GetPromItem(), command: ", command)

	out, err := utils.Cmd(command)
	if err != nil {
		return "", err
	}

	log.Println("+++ in GetPromItem(), out")
	log.Println(out)
	log.Println("--- out")

	err = json.Unmarshal([]byte(out), &rsp)
	if err != nil {
		return "", err
	}
	if rsp.Status != config.STATUS_SUCCCESS {
		return "", err
	}

	for _, v := range rsp.Data.Result {
		if v.Metric.Instance == host {
			for _, v2 := range v.Value {
				varType := reflect.TypeOf(v2) //float64: timestamp, string: cpu usage
				switch varType.Name() {
				case "string":
					return v2.(string), nil
				}
			}
		}
	}

	return "", nil
}

func GetPromRange(promType string, host string, start int, end int, step int) (*string, error) {
	var command string
	var rsp Response
	var out string
	var err error

	var promWebHost = utils.PromServer + ":" + utils.PromPort
	url := "http://" + promWebHost + "/api/v1/query_range?query="

	if promType == "disk" {
		promStr := "100%20-%20(node_filesystem_free_bytes{mountpoint=\"/\",fstype=~\"ext4|xfs\"}%20/%20node_filesystem_size_bytes{mountpoint=\"/\",fstype=~\"ext4|xfs\"}%20*%20100)"
		command = "curl -g '" + url + promStr +
			"&start=" + strconv.Itoa(start) +
			"&end=" + strconv.Itoa(end) +
			"&step=" + strconv.Itoa(step) + "s' 2>/dev/null"
		out, err = utils.Cmd(command)
		if err != nil {
			log.Println("curl error: ", err)
			return nil, err
		}
	}

	if promType == "mem" {
		promStr := "(node_memory_MemFree_bytes%2Bnode_memory_Cached_bytes%2Bnode_memory_Buffers_bytes)%20/%20node_memory_MemTotal_bytes%20*%20100"
		command = "curl -g '" + url + promStr +
			"&start=" + strconv.Itoa(start) +
			"&end=" + strconv.Itoa(end) +
			"&step=" + strconv.Itoa(step) + "s' 2>/dev/null"
		out, err = utils.Cmd(command)
		if err != nil {
			log.Println("curl error: ", err)
			return nil, err
		}
	}

	if promType == "cpu" {
		//curl -i -g 'http://10.0.0.24:9090/api/v1/query_range?query=100%20-%20(avg(irate(node_cpu_seconds_total{mode="idle"}[5m]))%20by%20(instance)%20*%20100)&start=1579150980&end=1585886888&step=2222s'
		promStr := "100%20-%20(avg(irate(node_cpu_seconds_total{mode=\"idle\"}[5m]))%20by%20(instance)%20*%20100)"
		command = "curl -g '" + url + promStr +
			"&start=" + strconv.Itoa(start) +
			"&end=" + strconv.Itoa(end) +
			"&step=" + strconv.Itoa(step) + "s' 2>/dev/null"
		out, err = utils.Cmd(command)

		log.Println("+++ in GetPromRange(), out, command: ", command)
		log.Println(out)
		log.Println("--- out")
		if err != nil {
			return nil, err
		}

	}
	if promType == "qps" || promType == "querys" || promType == "memhit" || promType == "recurquerys" || promType == "dhcppacket" {
		//url := "http://10.0.0.24:9090/api/v1/query_range?query=dns_gauge%7Bdata_type%3D%22qps%22%2Cinstance%3D%2210.0.0.19%3A8001%22%7D&start=1582636272.047&end=1582639872.047&step=14"
		client := &http.Client{}
		var url string

		if promType == "qps" || promType == "dhcppacket" || promType == "dhcplease" {
			url = "http://" + promWebHost + "/api/v1/query_range?query=dns_gauge%7Bdata_type%3D%22" + promType + "%22%2Cinstance%3D%22" + host + "%3A8001%22%7D&start=" + strconv.Itoa(start) + "&end=" + strconv.Itoa(end) + "&step=" + strconv.Itoa(step)
		} else if promType == "querys" || promType == "memhit" || promType == "recurquerys" {
			url = "http://" + promWebHost + "/api/v1/query_range?query=dns_counter%7Bdata_type%3D%22" + promType + "%22%2Cinstance%3D%22" + host + "%3A8001%22%7D&start=" + strconv.Itoa(start) + "&end=" + strconv.Itoa(end) + "&step=" + strconv.Itoa(step)
		}
		log.Println("url: ", url)
		reqest, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		response, _ := client.Do(reqest)
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		out = string(body)
		if err != nil {
			return nil, err
		}

	}

	log.Println("GetPromRange(), out: ", out)
	d := json.NewDecoder(bytes.NewReader([]byte(out)))
	d.UseNumber()
	err = d.Decode(&rsp)

	if rsp.Status != config.STATUS_SUCCCESS {
		return nil, err
	}
	for _, v := range rsp.Data.Result {

		idx := strings.Index(v.Metric.Instance, ":")
		//log.Println("idx: ", idx)
		newInstance := v.Metric.Instance[:idx]
		if newInstance == host {

			retJson, err := json.Marshal(v.Values)
			if err != nil {
				log.Println("json marshal err: ", err)
			}
			//log.Println("string retJson: ", string(retJson))
			tmp := string(retJson)

			return &tmp, nil
		}
	}

	log.Println("return error")

	str := ""
	return &str, nil
}

func List_server(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	fmt.Println("in list_server(), Form: ", r.Form)

	//get servers maintained by kafka server
	result := NewBaseJsonServer()

	result.Status = config.STATUS_SUCCCESS
	result.Message = config.MSG_OK
	for _, s := range utils.OnlinePromHosts {
		result.Data = append(result.Data, s)
	}

	log.Println("+++ result")
	log.Println(result)
	log.Println("--- result")
	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))

	return
}
