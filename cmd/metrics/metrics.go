package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os/exec"
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type fooCollector struct {
	fooMetric *prometheus.Desc
	barMetric *prometheus.Desc
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func newFooCollector() *fooCollector {
	return &fooCollector{
		fooMetric: prometheus.NewDesc("fff_metric",
			"Shows whether a foo has occurred in our cluster",
			nil, nil,
		),
		barMetric: prometheus.NewDesc("bbb_metric",
			"Shows whether a bar has occurred in our cluster",
			nil, nil,
		),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *fooCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.fooMetric
	ch <- collector.barMetric
}

//Collect implements required collect function for all promehteus collectors
func (collector *fooCollector) Collect(ch chan<- prometheus.Metric) {

	//Implement logic here to determine proper metric value to return to prometheus
	//for each descriptor or call other functions that do so.
	var metricValue float64

	v := url.Values{}
	cpu := "100 - (avg(irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) by (instance) * 100)"
	v.Add("query", cpu)
	curlCmd := " curl 'http://10.0.0.23:9090/api/v1/query?" + v.Encode() + "'"
	ret, err := cmd(curlCmd)
	if err != nil {
		log.Println("curl error, ", curlCmd)
	}
	log.Println("+++ ret")
	log.Println(ret)
	log.Println("--- ret")

	v2 := url.Values{}
	mem := "(node_memory_MemFree_bytes+node_memory_Cached_bytes+node_memory_Buffers_bytes) / node_memory_MemTotal_bytes * 100"
	v2.Add("query", mem)
	curlCmd2 := "curl 'http://10.0.0.23:9090/api/v1/query?" + v2.Encode() + "'"
	ret2, err2 := cmd(curlCmd2)
	if err2 != nil {
		log.Println("curl mem error, ", curlCmd2)
	}
	log.Println("+++ ret2")
	log.Println(ret2)
	log.Println("--- ret2")
	if 1 == 1 {
		metricValue = 1
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.

	ch <- prometheus.MustNewConstMetric(collector.fooMetric, prometheus.CounterValue, metricValue)
	ch <- prometheus.MustNewConstMetric(collector.barMetric, prometheus.CounterValue, metricValue)

}

func main() {

	//Create a new instance of the foocollector and
	//register it with the prometheus client.
	foo := newFooCollector()
	prometheus.MustRegister(foo)

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :10001")
	log.Fatal(http.ListenAndServe(":10002", nil))
}

func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}
