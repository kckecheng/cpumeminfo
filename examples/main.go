package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kckecheng/osprobe/probe"
	"github.com/kckecheng/osprobe/probe/linux"
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

	type config struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
		Port     int    `json:"port"`
		Type     string `json:"type"`
	}

	var hostCfgs []config
	err = json.Unmarshal(jbytes, &hostCfgs)
	if err != nil {
		log.Fatal("Fail to decode json", err)
	}

	for _, cfg := range hostCfgs {
		var p probe.Probe
		switch t := cfg.Type; t {
		case "windows":
			p, err = windows.NewWinServer(cfg.Host, cfg.User, cfg.Password, cfg.Port)
		case "esxi":
			p, err = vmware.NewVMWServer(cfg.Host, cfg.User, cfg.Password, cfg.Port)
		case "linux":
			p, err = linux.NewLinServer(cfg.Host, cfg.User, cfg.Password, cfg.Port)
		default:
			log.Fatal("Not a supported operating system")
		}
		if err != nil {
			log.Fatal("Fail to initial object", cfg, err)
		}

		online := p.Online()
		if online {
			fmt.Println("Server is online")
		} else {
			fmt.Println("Server is not online")
			return
		}

		cusage, err := p.GetCPUUsage()
		if err != nil {
			log.Error(err)
		} else {
			fmt.Printf("CPU usage: %+v\n", cusage)
		}

		musage, err := p.GetMemUsage()
		if err != nil {
			log.Error(err)
		} else {
			fmt.Printf("Memory usage: %+v\n", musage)
		}

		dusage, err := p.GetLocalDiskUsage()
		if err != nil {
			log.Error(err)
		} else {
			fmt.Printf("Local disk usage: %+v\n", dusage)
		}

		nusage, err := p.GetNICUsage()
		if err != nil {
			log.Error(err)
		} else {
			fmt.Printf("NIC usage: %+v\n", nusage)
		}
	}
}
