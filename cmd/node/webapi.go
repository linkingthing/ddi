package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

type BaseJsonBean struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func NewBaseJsonBean() *BaseJsonBean {
	return &BaseJsonBean{}
}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}

func query_range(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("in query_range() Form: ", r.Form)

}
func query(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	fmt.Println("Form: ", r.Form)
	//for k, v := range r.Form {
	//	fmt.Println(k, "=>", v, strings.Join(v, "-"))
	//}
	host, flagHost := r.Form["host"]
	promType, flagType := r.Form["promType"]
	if !(flagHost && flagType) {

		fmt.Fprint(w, "host and promType error")
		return
	}
	log.Println("host: ", host)
	log.Println("promType: ", promType)

	result := NewBaseJsonBean()

	result.Code = 200
	result.Message = "host and promType ok"

	resp, err := GetPromItem("cpu", "10.0.0.15:9100")
	if err != nil {
		log.Println(err)
		return
	}

	cpuUsage, err := strconv.ParseFloat(resp, 64)
	resp = fmt.Sprintf("%.2f", cpuUsage)

	result.Data = resp
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
		command = "curl -H \"Content-Type: application/json\" + " +
			"http://10.0.0.23:9090/api/v1/query?query=instance:node_cpu:avg_rate5m 2>/dev/null"
	}
	out, err := cmd(command)
	if err != nil {
		return "", err
	}

	log.Println("+++ out")
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

func main() {
	//v1
	mux := http.NewServeMux()
	mux.Handle("/", &myHandler{})
	mux.HandleFunc("/query", query)
	mux.HandleFunc("/query_range", query_range)

	log.Println("Starting v2 httpserver")
	log.Fatal(http.ListenAndServe(":1210", mux))
}
