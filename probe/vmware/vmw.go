package vmware

/*
	Connect to ESXi - do not connect to vCenter although it works
*/

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/kckecheng/osprobe/probe"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

// VMWServer vCenter/ESXi
type VMWServer struct {
	probe.Server
	client *vim25.Client
}

var esxiHosts []mo.HostSystem

// NewVMWServer init
func NewVMWServer(host, user, password string, port int) (*VMWServer, error) {
	s := probe.Server{
		Host:     host,
		User:     user,
		Password: password,
		Port:     port,
	}
	if !s.Valid() {
		return nil, errors.New("Inputs are not valid, please check")
	}

	u := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/sdk",
	}
	u.User = url.UserPassword(user, password)

	ctx := context.Background()
	c, err := govmomi.NewClient(ctx, &u, true)
	if err != nil {
		log.Errorf("Fail to create client for %+v", s)
		return nil, err
	}

	serv := VMWServer{
		Server: s,
		client: c.Client,
	}
	return &serv, nil
}

// GetCPUUsage implement interface
func (vmw VMWServer) GetCPUUsage() (float64, error) {
	var err error
	if esxiHosts == nil || len(esxiHosts) == 0 {
		esxiHosts, err = vmw.getHostMor()
		if err != nil {
			return 0, err
		}
	}

	var ret []float64
	for _, h := range esxiHosts {
		totalCPU := int64(h.Summary.Hardware.CpuMhz) * int64(h.Summary.Hardware.NumCpuCores)
		usedCPU := int64(h.Summary.QuickStats.OverallCpuUsage)

		ret = append(ret, (float64(usedCPU)/float64(totalCPU))*100)
	}
	return ret[0], nil
}

// GetMemUsage implement interface
func (vmw VMWServer) GetMemUsage() (float64, error) {
	var err error
	if esxiHosts == nil || len(esxiHosts) == 0 {
		esxiHosts, err = vmw.getHostMor()
		if err != nil {
			return 0, err
		}
	}

	var ret []float64
	for _, h := range esxiHosts {
		totalMemory := int64(h.Summary.Hardware.MemorySize)
		usedMemory := (int64(h.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024)
		ret = append(ret, float64(usedMemory)/float64(totalMemory)*100)
	}
	return ret[0], nil
}

// GetNICUsage implement interface
func (vmw VMWServer) GetNICUsage() (map[string]map[string]float64, error) {
	return nil, errors.New("Not implemented")
}

// GetLocalDiskUsage implement interface
func (vmw VMWServer) GetLocalDiskUsage() (map[string]float64, error) {
	return nil, errors.New("Not implemented")
}

func (vmw VMWServer) getHostMor() ([]mo.HostSystem, error) {
	c := vmw.client
	m := view.NewManager(c)

	ctx := context.Background()
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		log.Errorf("Fail to create host view due to %s", err)
		return nil, err
	}

	var hosts []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hosts)
	if err != nil {
		log.Errorf("Fail to grab host summary information due to %s", err)
		return nil, err
	}
	return hosts, nil
}
