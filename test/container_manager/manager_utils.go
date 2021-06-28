package container_manager

import (
	"log"
	"net"
	"time"
)


func checkPortAvailability(host string, requiredPorts []string) (string, error)    {
	log.Println("checking port availability")
	for _, port := range requiredPorts  {
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			log.Println("Connecting error:", err)
			return port, err
		}
		if conn != nil {
			defer conn.Close()
			log.Println("Opened", net.JoinHostPort(host, port))
		}
	}
	return  "", nil
}