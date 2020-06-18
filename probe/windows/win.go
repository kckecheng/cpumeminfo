package windows

/*
Connect to Windows with WinRM and grab information by running PowerShell commands

Prequisites: refer to https://github.com/masterzen/winrm
	winrm quickconfig
	y
	winrm set winrm/config/service/Auth '@{Basic="true"}'
	winrm set winrm/config/service '@{AllowUnencrypted="true"}'
	winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="1024"}'
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/kckecheng/osprobe/probe"
	"github.com/masterzen/winrm"
	log "github.com/sirupsen/logrus"
)

// WinServer Windows object
type WinServer struct {
	probe.Server
	client *winrm.Client
}

// NewWinServer init a Windows connection
func NewWinServer(host string, user, password string, port int) (*WinServer, error) {
	endpoint := winrm.NewEndpoint(host, port, false, false, nil, nil, nil, 0)
	client, err := winrm.NewClient(endpoint, user, password)
	if err != nil {
		log.Errorf("Fail to create Windows connection for %s", host)
		return nil, err
	}
	serv := WinServer{
		Server: probe.Server{
			Host:     host,
			User:     user,
			Password: password,
			Port:     port,
		},
		client: client,
	}
	return &serv, nil
}

// GetCPUUsage implement interface
func (win WinServer) GetCPUUsage() ([]float32, error) {
	var ret []float32
	var err error

	cpups := "(Get-WmiObject win32_processor | Select-Object -Property LoadPercentage | ConvertTo-Json).ToString()"
	log.Infof("Run command %s to get CPU stats", cpups)
	output, err := win.runCmd(cpups)
	if err != nil {
		return nil, err
	}

	// negative return implies invalid input
	extractLoad := func(cpu map[string]float32) (float32, error) {
		load, exists := cpu["LoadPercentage"]
		if exists {
			return load, nil
		}
		return 0, fmt.Errorf("Key LoadPercentage does not exist within %+v", cpu)
	}

	if strings.HasPrefix(output, "[") {
		var cpus []map[string]float32
		err = json.Unmarshal([]byte(output), &cpus)
		if err != nil {
			log.Errorf("Fail to decode CPU stats %s with error %s", output, err)
			return nil, err
		}

		for _, cpu := range cpus {
			load, err := extractLoad(cpu)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			ret = append(ret, load)
		}
		return ret, nil
	} else if strings.HasPrefix(output, "{") {
		var cpu map[string]float32
		err = json.Unmarshal([]byte(output), &cpu)
		if err != nil {
			log.Errorf("Fail to decode CPU stats %s with error %s", output, err)
			return nil, err
		}

		load, err := extractLoad(cpu)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		return append(ret, load), nil
	}

	err = fmt.Errorf("The command output %s is not a valid json string", output)
	log.Error(err)
	return nil, err
}

// GetMemUsage implement interface
func (win WinServer) GetMemUsage() (float32, error) {
	memps := "Get-WmiObject win32_OperatingSystem | Select-Object -Property FreePhysicalMemory,TotalVisibleMemorySize | ConvertTo-Json"
	log.Infof("Run command %s to get memory stats", memps)
	output, err := win.runCmd(memps)
	if err != nil {
		return 0, err
	}

	var memory map[string]float32
	err = json.Unmarshal([]byte(output), &memory)
	if err != nil {
		log.Errorf("Fail to decode memory stats %s with error %s", output, err)
		return 0, err
	}

	free, exists := memory["FreePhysicalMemory"]
	if !exists {
		err = fmt.Errorf("Key FreePhysicalMemory does not exist within %+v", memory)
		log.Error(err)
		return 0, err
	}

	total, exists := memory["TotalVisibleMemorySize"]
	if !exists {
		err = fmt.Errorf("Key TotalVisibleMemorySize does not exist within %+v", memory)
		log.Error(err)
		return 0, err
	}

	return 1 - free/total, nil
}

// Only powershell command is supported
func (win WinServer) runCmd(cmd string) (string, error) {
	log.Infof("Execute command %s", cmd)

	var buf bytes.Buffer
	_, err := win.client.Run(winrm.Powershell(cmd), &buf, ioutil.Discard)
	if err != nil {
		log.Errorf("Fail to run command %s with error %s", cmd, err)
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
