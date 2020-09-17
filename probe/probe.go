package probe

import (
	"bytes"
	"os/exec"
	"strings"
)

// Probe interface
type Probe interface {
	Online() bool
	GetCPUUsage() (float64, error)
	GetMemUsage() (float64, error)
	GetLocalDiskUsage() (map[string]float64, error)
	GetNICUsage() (map[string]map[string]float64, error)
}

// Server inforamtion to connect to a server
type Server struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Type     string `json:"type"` // linux, windows, or esxi
}

// Valid make sure all fields are valid
func (s Server) Valid() bool {
	if s.Host == "" || s.User == "" || s.Password == "" || s.Port <= 0 || s.Port > 65535 {
		return false
	}

	switch s.Type {
	case "linux", "windows", "esxi":
		return true
	}
	return false
}

// Online check if a server is reachable
func (s Server) Online() bool {
	var out bytes.Buffer
	cmd := exec.Command("ping", "-W", "1", "-c", "3", s.Host)
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return false
	}

	if strings.Contains(out.String(), "100% packet loss") {
		return false
	}
	return true
}
