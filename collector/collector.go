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

// ServerCollector prometheus collector
type ServerCollector struct {
	Servers []probe.Server
	Stat    map[string]map[string]float64
	Mutex   sync.Mutex
}

// NewServerCollector init collector
func NewServerCollector(path string) *ServerCollector {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Fail to read file %s due to %s", path, err)
	}

	var servers []probe.Server
	err = json.Unmarshal(contents, &servers)
	if err != nil {
		log.Fatal("Fail to decode json", err)
	}
	for _, server := range servers {
		if server.Type != "linux" && server.Type != "windows" && server.Type != "esxi" {
			// log.Fatalf("Server type %s is not supported, please check the configuration", server.Type)
			log.Errorf("Server type %s is not supported (%s), please check the configuration", server.Type, server.Host)
		}
	}

	collector := ServerCollector{
		Servers: servers,
		Stat:    map[string]map[string]float64{},
		Mutex:   sync.Mutex{},
	}
	return &collector
}

// Describe implement prometheus collector required interface
func (sc *ServerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range descs {
		ch <- v
	}
}

// Collect implement prometheus collector required interface
func (sc *ServerCollector) Collect(ch chan<- prometheus.Metric) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	descKeys := []string{"online", "accessible", "cpu_utilization", "mem_utilization"}
	for k, v := range sc.Stat {
		target := sc.findServer(k)

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

func (sc *ServerCollector) findServer(host string) probe.Server {
	for _, s := range sc.Servers {
		if s.Host == host {
			return s
		}
	}
	return probe.Server{}
}
