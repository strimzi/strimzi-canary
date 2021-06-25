package container_manager

import (
	"github.com/docker/go-connections/nat"
	"log"
	"net"
	"regexp"
	"time"
)

func parsePortNumber( port nat.Port) string {
	r, _ := regexp.Compile("[^\\s](\\d{1,})")
	portResolved := r.FindString(string(port))
	return portResolved

}

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