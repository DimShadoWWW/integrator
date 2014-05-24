package vulcand

import (
	"errors"
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/mailgun/vulcand/backend"
	"github.com/mailgun/vulcand/backend/etcdbackend"
	"math/rand"
	"strconv"
	"time"
)

var (
	ipaddress string
)

type VulcandHostJSON struct {
	Host string ` json:"host"`
	Port int    `json:"port"`
}

func VulcandHostAdd(etcdnodes []string, dockeruri string, service fleet.SystemdService, port int, path string) error {

	rand.Seed(time.Now().UnixNano())

	fullname := service.Hostname + "." + service.Domain

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(service.Hostname + "-" + strconv.FormatInt(service.Id, 10))
		if err != nil {
			if i == 9 {
				return err
			}
		} else {
			break
		}
		// wait 10 seconds to try again
		time.Sleep(10 * time.Second)
	}

	client, err := etcdbackend.NewEtcdBackend(etcdnodes, "vulcand", "STRONG")
	if err != nil {
		return err
	}

	var host_exist backend.Host
	var endpointId, locationId, upstreamId string

	hosts, err := client.GetHosts()
	for _, host := range hosts {
		if host.Name == fullname {
			host_exist = *host
			fmt.Println("Hosts exist")
			break
		}
	}

	if host_exist.Name == "" {
		err = client.AddHost(fullname)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Checking hosts")
		for _, location := range host_exist.Locations {
			fmt.Println("Checking location", location.Hostname)
			if location.Path == path {
				locationId = location.Id
				upstreamId = location.Upstream.Id
				for _, endpoint := range location.Upstream.Endpoints {
					if endpoint.Url == "http://"+ipaddress+":"+strconv.Itoa(port) {
						endpointId = endpoint.Id
						break
						break
					}
				}
			}
		}
	}

	if upstreamId == "" {
		fmt.Println("createUpstream")
		upstreamId = genId()
		err = client.AddUpstream(upstreamId)
		if err != nil {
			return err
		}
	}

	if locationId == "" {
		fmt.Println("createLocation")
		locationId = strconv.FormatInt(service.Id, 10)
		err = client.AddLocation(locationId, fullname, path, upstreamId)
		if err != nil {
			return err
		}
	}

	if endpointId == "" {
		fmt.Println("createEndpoint")
		endpointId = strconv.FormatInt(service.Id, 10)
		err = client.AddEndpoint(upstreamId, endpointId, "http://"+ipaddress+":"+strconv.Itoa(port))
		if err != nil {
			return err
		}
	}

	return nil
}

func VulcandHostDel(etcdnodes []string, dockeruri string, service fleet.SystemdService, port int, path string) error {
	fullname := service.Hostname + "." + service.Domain
	fmt.Println(fullname)

	ipaddress := "127.0.0.1"

	client, err := etcdbackend.NewEtcdBackend(etcdnodes, "vulcand", "STRONG")
	if err != nil {
		return err
	}

	var host_exist backend.Host
	var endpointId, locationId, upstreamId string

	hosts, err := client.GetHosts()
	for _, host := range hosts {
		if host.Name == fullname {
			host_exist = *host
			fmt.Println("Hosts exist")
			break
		}
	}

	if host_exist.Name == "" {
		err = client.AddHost(fullname)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Checking hosts")
		for _, location := range host_exist.Locations {
			fmt.Println("Checking location", location.Hostname)
			if location.Path == path {
				locationId = location.Id
				upstreamId = location.Upstream.Id
				for _, endpoint := range location.Upstream.Endpoints {
					if endpoint.Url == "http://"+ipaddress+":"+strconv.Itoa(port) {
						endpointId = endpoint.Id
						fmt.Println(endpoint.Stats)
						break
						break
					}
				}
			}
		}
	}

	if endpointId == "" && upstreamId == "" && locationId == "" {
		return errors.New("Endpoint not found")
	}

	if endpointId != "" {
		fmt.Println("createEndpoint")
		err = client.DeleteEndpoint(upstreamId, endpointId)
		if err != nil {
			return err
		}
	}
	if locationId != "" {
		fmt.Println("createLocation")
		err = client.DeleteLocation(fullname, locationId)
		if err != nil {
			return err
		}
	}
	if upstreamId != "" {
		fmt.Println("createUpstream")
		err = client.DeleteUpstream(upstreamId)
		if err != nil {
			return err
		}
	}

	return nil
}

func genId() string {
	var id int64
	for i := 0; i < 10; i++ {
		id = rand.Int63() + 1
	}
	return string(strconv.FormatInt(id, 10))
}
