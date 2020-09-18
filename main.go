package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kckecheng/osprobe/collector"
	"github.com/kckecheng/osprobe/probe"
	"github.com/kckecheng/osprobe/probe/linux"
	"github.com/kckecheng/osprobe/probe/vmware"
	"github.com/kckecheng/osprobe/probe/windows"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

// INTERVAL update interval
const INTERVAL time.Duration = 10

func init() {
	// log.SetLevel(log.DebugLevel)
	// log.SetLevel(log.InfoLevel)
	log.Infof("Results will be updated every %d seconds", INTERVAL)
}

func deleteJob(pusher *push.Pusher, gateway, job string) {
	log.Debugf("Delete job %s from pushgateway %s", job, gateway)

	if err := pusher.Delete(); err != nil {
		log.Error("Fail to delete the Pushgateway job:", err)
		log.Print("Please delete the job manually as:", fmt.Sprintf("curl -X DELETE %s/metrics/job/%s", gateway, job))
	}
}

func main() {
	// Parse arguments
	var job, gateway, cfg string
	flag.StringVarP(&job, "job", "j", "osprobe", "Pushgateway job name")
	flag.StringVarP(&gateway, "gateway", "g", "http://127.0.0.1:9091", "Pushgateway URL")
	flag.StringVarP(&cfg, "config", "c", "hosts.json", "Host definitions")

	flag.Parse()
	if job == "" || gateway == "" || cfg == "" {
		flag.Usage()
		os.Exit(1)
	}
	log.Infof("Probe results will be pushed to %s with job %s", gateway, job)

	// Collector init and register
	sc := collector.NewSrvCollector(cfg)
	reg := prometheus.NewRegistry()
	reg.MustRegister(sc)

	// Pusher init
	pusher := push.New(gateway, job).Collector(sc)
	defer func() {
		deleteJob(pusher, gateway, job)
	}()

	// Register signal handler
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigc
		log.Infof("Signal captured, exit the application")

		deleteJob(pusher, gateway, job)
		defer os.Exit(1)
	}()

	// Mark if a round of probe is done
	pdone := make(chan int)
	// Probe based on defind interval in the background
	go func() {
		ticker := time.NewTicker(INTERVAL * time.Second)
		defer ticker.Stop()

		for {
			// Periodical probe over servers
			select {
			case <-ticker.C:
				var wg sync.WaitGroup
				for _, server := range sc.Servers {
					wg.Add(1)
					go func(server probe.Server) {
						log.Debug("Probe serve:", server.Host)
						defer wg.Done()

						// Initial stat for each server
						stat := map[string]float64{
							"online":          0,
							"accessible":      0,
							"cpu_utilization": 0,
							"mem_utilization": 0,
						}

						online := server.Online()
						if !online {
							log.Errorf("Server %s is offline", server.Host)
							sc.Mutex.Lock()
							sc.Stat[server.Host] = stat
							sc.Mutex.Unlock()
							return
						}
						stat["online"] = 1

						log.Debug("Create connection to server:", server.Host)
						var p probe.Probe
						var err error
						switch t := server.Type; t {
						case "linux":
							p, err = linux.NewLinServer(server.Host, server.User, server.Password, server.Port)
						case "windows":
							p, err = windows.NewWinServer(server.Host, server.User, server.Password, server.Port)
						case "esxi":
							p, err = vmware.NewVMWServer(server.Host, server.User, server.Password, server.Port)
						}

						if err != nil {
							log.Error("Fail to connect to server:", server.Host)
							sc.Mutex.Lock()
							sc.Stat[server.Host] = stat
							sc.Mutex.Unlock()
							return
						}
						stat["accessible"] = 1

						log.Debug("Gather CPU usage for server:", server.Host)
						cpuUsage, err := p.GetCPUUsage()
						if err != nil {
							log.Error("Fail to probe CPU usage", err)
						} else {
							stat["cpu_utilization"] = cpuUsage
						}

						log.Debug("Gather memory usage for server:", server.Host)
						memUsage, err := p.GetMemUsage()
						if err != nil {
							log.Error("Fail to probe memory usage", err)
						} else {
							stat["mem_utilization"] = memUsage
						}

						log.Debug("Update latest stat for server:", server.Host)
						sc.Mutex.Lock()
						sc.Stat[server.Host] = stat
						sc.Mutex.Unlock()
					}(server)
				}
				wg.Wait()
				// Complete one round of collection
				pdone <- 1
			}
		}
	}()

	// Push whenever a round of probe results is ready
	for {
		<-pdone
		if err := pusher.Push(); err != nil {
			log.Fatal("Fail to push metrics", err)
		}
		log.Info("Push 1 x round of probe results")
	}
}
