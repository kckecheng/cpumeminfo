package linux

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/kckecheng/osprobe/probe"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// LinServer Linux server
type LinServer struct {
	probe.Server
	client *ssh.Client
}

// NewLinServer init
func NewLinServer(host, user, password string, port int) (*LinServer, error) {
	s := probe.Server{
		Host:     host,
		User:     user,
		Password: password,
		Port:     port,
		Type:     "linux",
	}
	if !s.Valid() {
		return nil, errors.New("Inputs are not valid, please check")
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		log.Errorf("Fail to establish ssh connection to %s due to %s", host, err)
		return nil, err
	}

	server := LinServer{
		Server: s,
		client: client,
	}
	return &server, nil
}

// GetCPUUsage implement interface
func (lin LinServer) GetCPUUsage() (float64, error) {
	cmd := "head -n1 /proc/stat"

	output, err := lin.run(cmd)
	if err != nil {
		log.Errorf("Fail to query CPU usage: %s", err)
		return 0, err
	}
	r, _ := regexp.Compile(`\d+`)
	fields := r.FindAllString(output, -1)

	var values []float64
	var total float64
	for _, field := range fields {
		v, _ := strconv.ParseFloat(field, 64)
		total += v
		values = append(values, v)
	}
	return values[2] * 100 / total, nil
}

// GetMemUsage implement interface
func (lin LinServer) GetMemUsage() (float64, error) {
	cmd := "head -n2 /proc/meminfo"

	output, err := lin.run(cmd)
	if err != nil {
		log.Errorf("Fail to query memory usage: %s", err)
		return 0, err
	}

	r, _ := regexp.Compile(`\d+`)
	fields := r.FindAllString(output, -1)
	memTotal, _ := strconv.ParseFloat(fields[0], 64)
	memFree, _ := strconv.ParseFloat(fields[1], 64)
	return (memTotal - memFree) * 100 / memTotal, nil
}

// GetLocalDiskUsage implement interface
func (lin LinServer) GetLocalDiskUsage() (map[string]float64, error) {
	return nil, errors.New("Not implemented")
}

// GetNICUsage implement interface
func (lin LinServer) GetNICUsage() (map[string]map[string]float64, error) {
	return nil, errors.New("Not implemented")
}

func (lin LinServer) run(cmd string) (string, error) {
	session, err := lin.client.NewSession()
	if err != nil {
		log.Error("Fail to create session", err)
		return "", err
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		log.Errorf("Fail to run command %s due to %s", cmd, err)
		return "", err
	}
	return b.String(), nil
}
