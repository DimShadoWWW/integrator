package main

import (
	"fmt"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/DimShadoWWW/integrator/proxyctl/hipache"
	"github.com/DimShadoWWW/integrator/proxyctl/vulcand"
	"github.com/alecthomas/kingpin"
	"os"
	"strconv"
	"time"
)

var (
	app         = kingpin.New("proxyctl", "A command-line application to add and remove services from vulcand in CoreOS.")
	serverIP    = app.Flag("server", "Etcd address").Default("127.0.0.1").IP()
	redisAccess = app.Flag("redis", "Redis address").Default("redis://:@127.0.0.1:6379/0").String()
	docker      = app.Flag("docker", "docker uri for fleet").Default("unix:///var/run/docker.sock").String()
	proxytype   = app.Flag("type", "proxy software (hipache, vulcand)").Default("hipache").String()
	TTL         = app.Flag("ttl", "server ttl in DB ('10s', '5m', '300ms')").Default("10s").String()
	Id          = app.Flag("id", "id for service").Required().String()
	Name        = app.Flag("name", "docker container's name").Required().String()
	Hostname    = app.Flag("hostname", "hostname for service").Required().String()
	Domain      = app.Flag("domain", "domain for service").Required().String()
	Region      = app.Flag("region", "region for service").Required().String()
	Port        = app.Flag("port", "service's listenning port").Required().Int()
	Path        = app.Flag("path", "path to serve (\"/.*\")").Default("/.*").String()
	add         = app.Command("add", "Add new services.")
	del         = app.Command("del", "Remove a service.")
)

func main() {

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// Register user
	case "add":
		ttl, err := time.ParseDuration(*TTL)
		if err != nil {
			fmt.Printf("TTL parsing error: %s\n", err)
			os.Exit(1)
		}

		machines := []string{"http://" + serverIP.String() + ":4001"}

		fmt.Printf("%s\n", machines)

		id, err := strconv.ParseInt(*Id, 10, 64)
		if err != nil {
			fmt.Printf("id conversion error: %s\n", err)
		}

		f := fleet.SystemdService{
			Id:       id,
			Name:     *Name,
			Hostname: *Hostname,
			Domain:   *Domain,
			Region:   *Region,
			HttpPort: *Port,
		}

		switch {
		case *proxytype == "vulcand":
			err = vulcand.VulcandHostAdd(machines, *docker, f, *Port, *Path, ttl)
			if err != nil {
				fmt.Printf("Proxy addition failed: %s\n", err)
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		case *proxytype == "hipache":
			err = hipache.HostAdd(*redisAccess, *docker, f, *Port, *Path)
			if err != nil {
				fmt.Printf("Proxy addition failed: %s\n", err)
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		default:
			fmt.Println("Proxy type not supported")

		}
	case "del":
		machines := []string{"http://" + serverIP.String() + ":4001"}

		fmt.Printf("%s\n", machines)

		id, err := strconv.ParseInt(*Id, 10, 64)
		if err != nil {
			fmt.Printf("id conversion errorr: %s\n", err)
		}

		f := fleet.SystemdService{
			Id:       id,
			Name:     *Name,
			Hostname: *Hostname,
			Domain:   *Domain,
			Region:   *Region,
			HttpPort: *Port,
		}

		switch {
		case *proxytype == "vulcand":
			err = vulcand.VulcandHostDel(machines, *docker, f, *Port, *Path)
			if err != nil {
				fmt.Printf("Proxy deletion failed: %s\n", err)
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
		case *proxytype == "hipache":
			err = hipache.HostDel(*redisAccess, *docker, f, *Port, *Path)
			if err != nil {
				fmt.Printf("Proxy deletion failed: %s\n", err)
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
		default:
			fmt.Println("Proxy type not supported")
			os.Exit(3)
		}
	default:
		fmt.Println("Command not found")
		os.Exit(4)
	}

}
