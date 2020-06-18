package probe

// Probe interface
type Probe interface {
	GetCPUUsage() ([]float32, error)
	GetMemUsage() (float32, error)
}

// Server inforamtion to connect to a server
type Server struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}
