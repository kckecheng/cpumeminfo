package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kckecheng/osprobe/probe"
	"github.com/kckecheng/osprobe/probe/vmware"
	"github.com/kckecheng/osprobe/probe/windows"
	log "github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <path to config json>", os.Args[0])
	}
	fname := os.Args[1]

	jbytes, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}

	var config probe.Server
	err = json.Unmarshal(jbytes, &config)
	if err != nil {
		panic(err)
	}

	var p probe.Probe

	if strings.Contains(fname, "winidows") {
		p, err = windows.NewWinServer(config.Host, config.User, config.Password, config.Port)
	} else if strings.Contains(fname, "esxi") {
		p, err = vmware.NewVMWServer(config.Host, config.User, config.Password, config.Port)
	}
	if err != nil {
		log.Error(err)
	}

	cusage, err := p.GetCPUUsage()
	if err != nil {
		log.Error(err)
	}
	fmt.Printf("CPU usage: %+v\n", cusage)

	musage, err := p.GetMemUsage()
	if err != nil {
		log.Error(err)
	}
	fmt.Printf("Memory usage: %+v\n", musage)

	dusage, err := p.GetLocalDiskUsage()
	if err != nil {
		log.Error(err)
	}
	fmt.Printf("Local disk usage: %+v\n", dusage)

	nusage, err := p.GetNICUsage()
	if err != nil {
		log.Error(err)
	}
	fmt.Printf("NIC usage: %+v\n", nusage)

}
