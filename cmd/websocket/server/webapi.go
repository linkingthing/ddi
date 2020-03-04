package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/linkingthing/ddi/utils"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

var (
	promServer = "10.0.0.24:9090"
	host       = "10.0.0.15:9100"
)

func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}

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
type myHandler struct{}

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

func NewBaseJsonBean() *BaseJsonBean {
	return &BaseJsonBean{}
}
func NewBaseJsonRange() *BaseJsonRange {
	return &BaseJsonRange{}
}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	//cpuResp, err := GetPromRange("cpu", "10.0.0.15", 1579150980, 1579154580, 323)
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
	//if !(flagHost && flagType) {
	//
	//	// todo no host or promType, means front wants all info
	//	//fmt.Fprint(w, "host and promType error")
	//
	//	//return
	//}
	log.Println("host: ", host)
	log.Println("promType: ", promType)

	result := NewBaseJsonBean()

	result.Status = "success"
	result.Message = "ok"
	result.Data.Nodes = utils.OnlinePromHosts

	//if postHost != "" {
	//	for k, v := range utils.OnlinePromHosts {
	//		log.Println("+++ k")
	//		log.Println(k)
	//		log.Println("--- k")
	//
	//		if k != postHost {
	//			log.Println("k != hostname, continue")
	//			continue
	//		}
	//
	//		log.Println("+++ v")
	//		log.Println(v)
	//		log.Println("--- v")
	//	}
	//
	//}

	cpuResp, err := GetPromItem("cpu", "10.0.0.15:9100")
	if err != nil {
		log.Println(err)
		return
	}

	cpuUsage, err := strconv.ParseFloat(cpuResp, 64)
	cpuResp = fmt.Sprintf("%.2f", cpuUsage)
	//log.Println("cpuResp: ", cpuResp)

	var HostUsage = make(map[string]Usage)
	var Usage = Usage{}
	Usage.Cpu = cpuResp
	Usage.Mem = strconv.Itoa(rand.Intn(99)) + "." + strconv.Itoa(rand.Intn(99))
	Usage.Disk = strconv.Itoa(rand.Intn(99)) + "." + strconv.Itoa(rand.Intn(99))
	Usage.Qps = strconv.Itoa(rand.Intn(99)) + "." + strconv.Itoa(rand.Intn(99))
	postHost := ""
	if host != nil {
		postHost = host[0]
	}

	log.Println("postHost: ", postHost)
	if postHost != "" {

		HostUsage[postHost] = Usage
	} else {
		HostUsage["10.0.0.15"] = Usage
		HostUsage["10.0.0.24"] = Usage
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
	if promType == "cpu" {
		command = "curl -H \"Content-Type: application/json\"  " +
			"http://10.0.0.24:9090/api/v1/query?query=instance:node_cpu:avg_rate5m 2>/dev/null"
	}
	out, err := cmd(command)
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
	if rsp.Status != "success" {
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

	url := "http://10.0.0.24:9090/api/v1/query_range?query="

	if promType == "disk" {
		promStr := "100%20-%20(node_filesystem_free_bytes{mountpoint=\"/\",fstype=~\"ext4|xfs\"}%20/%20node_filesystem_size_bytes{mountpoint=\"/\",fstype=~\"ext4|xfs\"}%20*%20100)"
		command = "curl -g '" + url + promStr +
			"&start=" + strconv.Itoa(start) +
			"&end=" + strconv.Itoa(end) +
			"&step=" + strconv.Itoa(step) + "s' 2>/dev/null"
		out, err = cmd(command)
		log.Println("+++ in GetPromRange(), out")
		log.Println(out)
		log.Println("--- out")
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
		out, err = cmd(command)
		log.Println("+++ in GetPromRange(), out")
		log.Println(out)
		log.Println("--- out")
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
		out, err = cmd(command)

		//log.Println("+++ in GetPromRange(), out")
		//log.Println(out)
		//log.Println("--- out")
		if err != nil {
			return nil, err
		}

	}
	if promType == "qps" || promType == "querys" || promType == "memhit" || promType == "recurquerys" {
		//url := "http://10.0.0.24:9090/api/v1/query_range?query=dns_gauge%7Bdata_type%3D%22qps%22%2Cinstance%3D%2210.0.0.19%3A8001%22%7D&start=1582636272.047&end=1582639872.047&step=14"
		client := &http.Client{}
		var url string
		if promType == "qps" {
			url = "http://10.0.0.24:9090/api/v1/query_range?query=dns_gauge%7Bdata_type%3D%22" + promType + "%22%2Cinstance%3D%22" + host + "%3A8001%22%7D&start=" + strconv.Itoa(start) + "&end=" + strconv.Itoa(end) + "&step=" + strconv.Itoa(step)
		} else if promType == "querys" || promType == "memhit" || promType == "recurquerys" {
			url = "http://10.0.0.24:9090/api/v1/query_range?query=dns_counter%7Bdata_type%3D%22" + promType + "%22%2Cinstance%3D%22" + host + "%3A8001%22%7D&start=" + strconv.Itoa(start) + "&end=" + strconv.Itoa(end) + "&step=" + strconv.Itoa(step)
		}
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
		//log.Println("+++ in GetPromRange(), out")
		//log.Println(out)
		//log.Println("--- out")
		if err != nil {
			return nil, err
		}

	}

	d := json.NewDecoder(bytes.NewReader([]byte(out)))
	d.UseNumber()
	err = d.Decode(&rsp)

	if rsp.Status != "success" {
		return nil, err
	}
	for _, v := range rsp.Data.Result {

		idx := strings.Index(v.Metric.Instance, ":")
		log.Println("idx: ", idx)
		newInstance := v.Metric.Instance[:idx]
		if newInstance == host {

			retJson, err := json.Marshal(v.Values)
			if err != nil {
				log.Println("json marshal err: ", err)
			}
			log.Println("string retJson: ", string(retJson))
			tmp := string(retJson)

			return &tmp, nil
		}
	}

	log.Println("return error")
	str := ""
	//str := `[[1579167980.752,"0.8333333341094402"],[1579168008.752,"0.7999999999689607"],[1579168036.752,"0.7999999999689607"],[1579168064.752,"0.8666666666977108"],[1579175176.752,"0.7999999999689607"]]`
	return &str, nil

}
