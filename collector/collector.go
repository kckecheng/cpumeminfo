package collector

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/kckecheng/osprobe/probe"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var descs = map[string]*prometheus.Desc{
	"online": prometheus.NewDesc(
		"online",
		"if the server is online based on icmp check: 1 - online, 0 - offline",
		[]string{"host", "type"},
		nil,
	),
	"accessible": prometheus.NewDesc(
		"accessible",
		"if the server can be logged in with the configured credential: 1 - accessible, 0 - not accessible ",
		[]string{"host", "type"},
		nil,
	),
	"cpu_utilization": prometheus.NewDesc(
		"cpu_utilization",
		"cpu utilization in percent",
		[]string{"host", "type"},
		nil,
	),
	"mem_utilization": prometheus.NewDesc(
		"mem_utilization",
		"memory utiliztion in percent",
		[]string{"host", "type"},
		nil,
	),
}

// SrvCollector prometheus collector
type SrvCollector struct {
	servers []probe.Server
	stat    map[string]map[string]float64
	mutex   sync.Mutex
}

// NewSrvCollector init collector
func NewSrvCollector(path string) *SrvCollector {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Fail to read file %s due to %s", path, err)
	}

	var servers []probe.Server
	err = json.Unmarshal(contents, &servers)
	if err != nil {
		log.Fatal("Fail to decode json", err)
	}

	collector := SrvCollector{
		servers: servers,
		stat:    map[string]map[string]float64{},
		mutex:   sync.Mutex{},
	}
	return &collector
}

// Describe implement prometheus collector required interface
func (sc *SrvCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range descs {
		ch <- v
	}
}

// Collect implement prometheus collector required interface
func (sc *SrvCollector) Collect(ch chan<- prometheus.Metric) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	descKeys := []string{"online", "accessible", "cpu_utilization", "mem_utilization"}
	for k, v := range sc.stat {
		target := sc.findSrv(k)

		for _, dk := range descKeys {
			ch <- prometheus.MustNewConstMetric(
				descs[dk],
				prometheus.GaugeValue,
				v[dk],
				target.Host,
				target.Type,
			)
		}
	}
}

func (sc *SrvCollector) findSrv(host string) probe.Server {
	for _, s := range sc.servers {
		if s.Host == host {
			return s
		}
	}
	return probe.Server{}
}
