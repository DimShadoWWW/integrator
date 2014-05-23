package main

import (
	"flag"
	"fmt"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/DimShadoWWW/integrator/proxyctl/gogeta"
	"github.com/DimShadoWWW/integrator/proxyctl/vulcand"
	"os"
	"strings"
)

func main() {
	proxytype := flag.String("type", "vulcand", "proxy software (gogeta, vulcand)")
	path := flag.String("path", "/.*", "proxy software (gogeta, vulcand)")
	machines := flag.String("endpoint", "http://127.0.0.1:4001", "etcd endpoint for fleet")
	docker := flag.String("docker", "unix:///var/run/docker.sock", "docker uri for fleet")

	command := flag.String("cmd", "add", "Action (add, del)")
	id := flag.Int64("id", 0, "Id")
	hostname := flag.String("hostname", "", "Hostname")
	domain := flag.String("domain", "local", "Domain")
	region := flag.String("region", "east", "Region")
	port := flag.Int("port", 80, "Port")

	flag.Parse()

	if *id != 0 || *hostname != "" {
		client, err := etcdlib.NewEtcdClient(strings.Split(*machines, ","))
		if err != nil {
			fmt.Printf("Etcd: %s\n", err)
		}

		f := fleet.SystemdService{
			Id:       *id,
			Hostname: *hostname,
			Domain:   *domain,
			Region:   *region,
			HttpPort: *port,
		}

		switch {
		case *command == "add" || *command == "Add":
			switch {
			case *proxytype == "vulcand":
				err = vulcand.VulcandHostAdd(client, *docker, f, *port, *path)
				if err != nil {
					fmt.Printf("Proxy addition failed: %s\n", err)
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			case *proxytype == "gogeta":
				err = gogeta.GogetaHostAdd(client, *docker, f, *port)
				if err != nil {
					fmt.Printf("Proxy addition failed: %s\n", err)
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			default:
				fmt.Println("Proxy type not supported")

			}
		case *command == "del" || *command == "Del":
			switch {
			case *proxytype == "vulcand":
				err = vulcand.VulcandHostDel(client, *docker, f, *port)
				if err != nil {
					fmt.Printf("Proxy deletion failed: %s\n", err)
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			case *proxytype == "gogeta":
				err = gogeta.GogetaHostDel(client, f)
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
	} else {
		flag.Usage()
	}
}
