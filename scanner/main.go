package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kckecheng/osprobe/probe"
	"github.com/kckecheng/osprobe/probe/linux"
	"github.com/kckecheng/osprobe/probe/vmware"
	"github.com/kckecheng/osprobe/probe/windows"
	flag "github.com/spf13/pflag"
)

// const definitions
const (
	ERREXIT               = 1
	TIMEOUT time.Duration = 3
)

func scanTPort(host string, port int) bool {
	con, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), TIMEOUT*time.Second)
	if err != nil {
		return false
	}
	con.Close()
	return true
}

/*
	Detect the most likely OS based on opened ports:
	- VMware:
		- 902: vSphere client to VM for ESXi 5.x and later
		- 903: vSphere client to VM for ESXi 3.5 and 4.x
		- 443: vSphere client to ESX/ESXi
	- Windows:
		- 445: Windows sharing
		- 3389: RDP
		- 5985: WinRM HTTP
		- 5986: WinRM HTTPS
	- Linux:
		- 22: SSH
*/
func guessOS(host string) string {
	if scanTPort(host, 443) && (scanTPort(host, 902) || scanTPort(host, 903)) {
		return "esxi"
	}

	if scanTPort(host, 3389) && (scanTPort(host, 5985) || scanTPort(host, 5986) || scanTPort(host, 445)) {
		return "windows"
	}

	if scanTPort(host, 22) {
		return "linux"
	}

	return "unknown"
}

/*
	cbd: credentails map as below
	{
		"linux": ["user1:password1", "user2:password2", ...],
		"windows": [...],
		"esxi": [...],
	}
*/
func matchCredential(server probe.Server, cdb map[string][]string) (string, string) {
	for _, upcomb := range cdb[server.Type] {
		up := strings.Split(upcomb, ":")

		var p probe.Probe
		var err error
		switch server.Type {
		case "linux":
			p, err = linux.NewServer(server.Host, up[0], up[1], server.Port)
		case "windows":
			p, err = windows.NewServer(server.Host, up[0], up[1], server.Port)
		case "esxi":
			p, err = vmware.NewServer(server.Host, up[0], up[1], server.Port)
		}

		if err != nil {
			continue
		}

		_, err = p.GetCPUUsage()
		if err != nil {
			continue
		} else {
			return up[0], up[1]
		}
	}

	return "", ""
}

/*
	host: IP/FQDN
	cdbpath: credentials json file with a format as below
	{
		"linux": ["user1:password1", "user2:password2", ...],
		"windows": [...],
		"esxi": [...],
	}
*/
func fillServer(host, cdbpath string) probe.Server {
	server := probe.Server{
		Host: host,
	}

	osType := guessOS(host)
	server.Type = osType

	switch osType {
	case "linux":
		server.Port = 22
	case "windows":
		server.Port = 5985
	case "esxi":
		server.Port = 443
	}

	contents, err := ioutil.ReadFile(cdbpath)
	if err != nil {
		panic(err)
	}

	var cdb map[string][]string
	err = json.Unmarshal(contents, &cdb)
	if err != nil {
		panic(err)
	}

	user, password := matchCredential(server, cdb)
	server.User = user
	server.Password = password
	return server
}

func main() {
	// Parse CLI options
	var hfpath, cfpath, ofpath string
	flag.StringVarP(&hfpath, "server", "s", "hosts.json", "Host IP/FQDB definitions json")
	flag.StringVarP(&cfpath, "password", "p", "credentials.json", "Credential database json")
	flag.StringVarP(&ofpath, "output", "o", "servers.json", "Output json")
	flag.Parse()

	if hfpath == "" || cfpath == "" {
		flag.Usage()
		os.Exit(ERREXIT)
	}

	// Parse host IP/FQDN definitions
	contents, err := ioutil.ReadFile(hfpath)
	if err != nil {
		panic(err)
	}

	var hosts []string
	err = json.Unmarshal(contents, &hosts)
	if err != nil {
		panic(err)
	}

	var servers []probe.Server
	var mutex sync.Mutex
	var wg sync.WaitGroup
	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			server := fillServer(host, cfpath)
			mutex.Lock()
			servers = append(servers, server)
			mutex.Unlock()
		}(host)
	}
	wg.Wait()

	bytes, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(ofpath, bytes, 0644)
	if err != nil {
		panic(err)
	}
}
