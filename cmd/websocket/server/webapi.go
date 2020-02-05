package main

import (
	"encoding/json"
	"fmt"
	"github.com/linkingthing/ddi/utils"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os/exec"
	"reflect"
	"strconv"
)

var (
	promServer = "10.0.0.23:9090"
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

type ValueIntf interface {
}

type Result struct {
	Metric Metric
	Value  []ValueIntf
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

func query_range(w http.ResponseWriter, r *http.Request) {
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
	err = json.Unmarshal([]byte(cpuResp), &histData)
	if err != nil {
		log.Println("cpuResp unmarshal error ", err)
	}

	result.Data.Values = histData

	//cpuHist, err := strconv.ParseFloat(cpuResp, 64)
	//cpuResp = fmt.Sprintf("%.2f", cpuUsage)
	log.Println("xxx cpuHist: ", cpuResp)

	bytes, _ := json.Marshal(result)
	//fmt.Fprint(w, string(bytes))
	w.Write([]byte(bytes))
}
func query(w http.ResponseWriter, r *http.Request) {

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
		HostUsage["10.0.0.23"] = Usage
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
			"http://10.0.0.23:9090/api/v1/query?query=instance:node_cpu:avg_rate5m 2>/dev/null"
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

func GetPromRange(promType string, host string, start int, end int, step int) (string, error) {
	var command string
	var rsp Response

	query := url.Values{}
	query.Add("query", "100 - (avg(irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) by (instance) * 100)")
	query.Add("start", "1579175189.588")
	query.Add("end", "1579178789.588")
	query.Add("step", "14")
	fmt.Printf("query str: %s", query.Encode())

	resUri, err := url.ParseRequestURI("query=&start=&end=&step=14")
	log.Println("resUri: ", resUri)

	str := `[[1579167980.752,"0.8333333341094402"],[1579168008.752,"0.7999999999689607"],[1579168036.752,"0.7999999999689607"],[1579168064.752,"0.8666666666977108"],[1579175176.752,"0.7999999999689607"]]`
	return str, nil

	urlPath := resUri.Query().Encode()
	if promType == "cpu" {
		command = "curl -H \"Content-Type: application/json\" " + "http://10.0.0.23:9090/api/v1/query_range" + urlPath + " 2>/dev/null"
	}
	log.Println("xxx command: ", command)
	out, err := cmd(command)
	if err != nil {
		return "", err
	}

	log.Println("+++ in GetPromRange(), out")
	log.Println(out)
	log.Println("--- out")

	//fmt.Println("promType, host, start, end, step,", promType, host, start, end, step)
	//var urlStr string = "http://baidu.com/index.php/?abc= 1_羽毛"
	//l, err := url.ParseRequestURI(urlStr)
	//fmt.Println(l, err)

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
