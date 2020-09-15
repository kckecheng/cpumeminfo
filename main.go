package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/kckecheng/osprobe/probe"
	"github.com/kckecheng/osprobe/probe/windows"
	log "github.com/sirupsen/logrus"
)

func main() {
	jbytes, err := ioutil.ReadFile("host.json")
	if err != nil {
		panic(err)
	}

	var config probe.Server
	err = json.Unmarshal(jbytes, &config)
	if err != nil {
		panic(err)
	}

	var p probe.Probe
	p, err = windows.NewWinServer(config.Host, config.User, config.Password, config.Port)
	if err != nil {
		log.Fatal(err)
	}

	cusage, err := p.GetCPUUsage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("CPU usage: %+v\n", cusage)

	musage, err := p.GetMemUsage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Memory usage: %+v\n", musage)

	dusage, err := p.GetLocalDiskUsage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Local disk usage: %+v\n", dusage)

	nusage, err := p.GetNICUsage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("NIC usage: %+v\n", nusage)

}
