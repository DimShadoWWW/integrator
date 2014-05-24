package main

import (
	"fmt"
	"github.com/DimShadoWWW/integrator/dnslib"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"github.com/alecthomas/kingpin"
	"os"
	"time"
)

var (
	// debug    = app.Flag("debug", "enable debug mode").Default("false").Bool()
	app      = kingpin.New("dnsctl", "A command-line application to add and remove services/$1 from vulcand in CoreOS.")
	serverIP = app.Flag("server", "Etcd address").Default("127.0.0.1").MetaVarFromDefault().IP()
	docker   = app.Flag("docker", "docker uri").Default("unix:///var/run/docker.sock").String()
	id       = app.Flag("id", "id for service").Required().Int64()
	hostname = app.Flag("hostname", "hostname for service").Required().String()
	domain   = app.Flag("domain", "domain for service").Required().String()
	region   = app.Flag("region", "region for service").Required().String()
	port     = app.Flag("port", "service's listenning port").Required().Int()
	priority = app.Flag("priority", "service's priority").Default("10").Int()
	add      = app.Command("add", "Generate and load new services.")
	del      = app.Command("del", "Unload a service.")
)

func main() {
	machines := []string{"http://" + string(*serverIP) + ":4001"}
	client, err := etcdlib.NewEtcdClient(machines)
	if err != nil {
		fmt.Printf("Etcd: %s\n", err)
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "add":
		for {
			err := dnslib.AddHostnameDNS(client, *docker, *id, *hostname, *domain, *port, *region, *priority)
			if err != nil {

			}
			time.Sleep(10 * time.Second)
		}
	case "del":
		dnslib.DeleteHostnameDNS(client, *id, *hostname, *domain, *port, *region)
	default:
		fmt.Println("Command not found")
		os.Exit(4)
	}
}
