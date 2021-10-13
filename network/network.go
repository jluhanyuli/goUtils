package network
import (
	"fmt"
	"os"
)

type Network struct {
	Ip   string
	Port string
	Net  string
}

const (
	EnvPort = "PORT"
	EnvIp   = "MY_HOST_IP"
)

func GetEnvWithDefault(key string, defaultValue string) string {
	if os.Getenv(key) != "" {
		return os.Getenv(key)
	} else {
		return defaultValue
	}
}


func NewNetwork() *Network {
	return &Network{
		Port: GetEnvWithDefault(EnvPort, "8080"),
		Ip:   GetEnvWithDefault(EnvIp, "127.0.0.1"),
		Net:  "tcp",
	}
}

func (n *Network) Network() string {
	return n.Net
}
func (n *Network) String() string {
	return fmt.Sprintf("%s:%s", n.Ip, n.Port)
}
