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
	"errors"
	"io/ioutil"
	"strings"

	"github.com/kckecheng/osprobe/probe"
	"github.com/masterzen/winrm"
	log "github.com/sirupsen/logrus"
)

// Server Windows object
type Server struct {
	probe.Server
	client *winrm.Client
}

// NewServer init a Windows connection
func NewServer(host, user, password string, port int) (Server, error) {
	server := Server{
		Server: probe.Server{
			Host:     host,
			User:     user,
			Password: password,
			Port:     port,
			Type:     "esxi",
		},
	}
	if !server.Valid() {
		return server, errors.New("Inputs are not valid, please check")
	}

	endpoint := winrm.NewEndpoint(host, port, false, false, nil, nil, nil, 0)
	client, err := winrm.NewClient(endpoint, user, password)
	if err != nil {
		log.Errorf("Fail to create client for %+v due to %s", server.Server, err)
		return server, err
	}

	server.client = client
	return server, nil
}

// GetCPUUsage implement interface
func (win Server) GetCPUUsage() (float64, error) {
	var ret float64

	cmd := "(Get-WmiObject win32_processor | Select-Object -Property DeviceID,LoadPercentage | ConvertTo-Json).ToString()"

	type cpuStats struct {
		ID   string  `json:"DeviceID"`
		Load float64 `json:"LoadPercentage"`
	}

	var cpus []cpuStats
	err := win.extractStats(cmd, &cpus)
	if err != nil {
		return ret, err
	}

	var total float64
	for _, cpu := range cpus {
		total += cpu.Load
	}
	ret = total / float64(len(cpus))
	return ret, nil
}

// GetMemUsage implement interface
func (win Server) GetMemUsage() (float64, error) {
	var ret float64

	cmd := "Get-WmiObject win32_OperatingSystem | Select-Object -Property FreePhysicalMemory,TotalVisibleMemorySize | ConvertTo-Json"

	type memStats struct {
		Free  float64 `json:"FreePhysicalMemory"`
		Total float64 `json:"TotalVisibleMemorySize"`
	}

	var mems []memStats
	err := win.extractStats(cmd, &mems)
	if err != nil {
		return ret, err
	}
	return (1 - mems[0].Free/mems[0].Total) * 100, nil
}

// GetLocalDiskUsage implement interface
func (win Server) GetLocalDiskUsage() (map[string]float64, error) {
	ret := map[string]float64{}

	cmd := "(Get-WmiObject -Class Win32_logicaldisk -Filter DriveType=3 | Select-Object -Property DeviceID,FreeSpace,Size | ConvertTo-Json).ToString()"

	type diskStats struct {
		DeviceID  string  `json:"DeviceID"`
		FreeSpace float64 `json:"FreeSpace"`
		Size      float64 `json:"Size"`
	}

	var disks []diskStats
	err := win.extractStats(cmd, &disks)
	if err != nil {
		return nil, err
	}

	for _, disk := range disks {
		did := strings.TrimRight(disk.DeviceID, ":")
		ret[did] = ((disk.Size - disk.FreeSpace) / disk.Size) * 100
	}
	return ret, nil
}

// GetNICUsage implement interface
func (win Server) GetNICUsage() (map[string]map[string]float64, error) {
	ret := map[string]map[string]float64{}

	cmd := "(Get-NetAdapterStatistics | Select-Object -Property Name,ReceivedBytes,SentBytes | ConvertTo-Json).ToString()"

	type nicStats struct {
		Name     string  `json:"Name"`
		Received float64 `json:"ReceivedBytes"`
		Sent     float64 `json:"SentBytes"`
	}

	var nics []nicStats
	err := win.extractStats(cmd, &nics)
	if err != nil {
		return nil, err
	}

	for _, nic := range nics {
		adapter := map[string]float64{}
		adapter["sent"] = nic.Sent
		adapter["received"] = nic.Received
		ret[nic.Name] = adapter
	}
	return ret, nil
}

func (win Server) extractStats(cmd string, stats interface{}) error {
	output, err := win.runCmd(cmd)
	if err != nil {
		return err
	}

	if strings.HasPrefix(output, "{") {
		output = "[" + output + "]"
	}

	// why below 2 x line of codes work the same???
	// err = json.Unmarshal([]byte(output), &stats)
	err = json.Unmarshal([]byte(output), stats)
	if err != nil {
		log.Errorf("Fail to extract stats %s with error %s", output, err)
		return err
	}
	return nil
}

// Only powershell command is supported
func (win Server) runCmd(cmd string) (string, error) {
	log.Debugf("Execute command %s", cmd)

	var buf bytes.Buffer
	_, err := win.client.Run(winrm.Powershell(cmd), &buf, ioutil.Discard)
	if err != nil {
		log.Errorf("Fail to run command %s with error %s", cmd, err)
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
