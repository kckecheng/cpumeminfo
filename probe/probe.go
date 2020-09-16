package probe

// Probe interface
type Probe interface {
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
}

// Valid make sure all fields are valid
func (s Server) Valid() bool {
	if s.Host == "" || s.User == "" || s.Password == "" || s.Port <= 0 || s.Port > 65535 {
		return false
	}
	return true
}
