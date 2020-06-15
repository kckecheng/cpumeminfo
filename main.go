package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/masterzen/winrm"
)

func main() {
	jbytes, err := ioutil.ReadFile("host.json")
	if err != nil {
		panic(err)
	}

	var config map[string]string
	err = json.Unmarshal(jbytes, &config)
	if err != nil {
		panic(err)
	}

	endpoint := winrm.NewEndpoint(config["host"], 5985, false, false, nil, nil, nil, 0)
	client, err := winrm.NewClient(endpoint, config["user"], config["password"])
	if err != nil {
		fmt.Println(err)
	}

	cpups := "(Get-WmiObject win32_processor | Select-Object -Property LoadPercentage | ConvertTo-Json).ToString()"
	memps := "Get-WmiObject win32_OperatingSystem | Select-Object -Property FreePhysicalMemory,TotalVisibleMemorySize | ConvertTo-Json"

	var buf bytes.Buffer
	_, err = client.Run(winrm.Powershell(cpups), &buf, ioutil.Discard)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(buf.String())

	buf.Reset()
	_, err = client.Run(winrm.Powershell(memps), &buf, ioutil.Discard)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(buf.String())
}
