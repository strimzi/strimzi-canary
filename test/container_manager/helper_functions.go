package container_manager

import (
	"github.com/docker/go-connections/nat"
	"regexp"
)

func parsePortNumber( port nat.Port) string {
	r, _ := regexp.Compile("[^\\s](\\d{1,})")
	portResolved := r.FindString(string(port))
	return portResolved

}