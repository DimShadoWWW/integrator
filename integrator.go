package main

import (
	"flag"
)

var (
	address string
)

func main() {
	// parse command line flags
	flag.StringVar(&address, "address", "unix:///var/run/docker.sock", "docker address")
	flag.Parse()

	client := NewDockerLib(address)
	containers := client.CleanContainers()
	client.RemoveContainers(containers)
}
