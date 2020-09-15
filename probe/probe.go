package probe

// Probe interface
type Probe interface {
	GetCPUUsage() (map[string]float64, error)
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
}
