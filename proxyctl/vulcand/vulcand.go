package vulcand

import (
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/mailgun/vulcand/backend"
	"github.com/mailgun/vulcand/backend/etcdbackend"
	"github.com/mailgun/vulcand/plugin/registry"
	"math/rand"
	"strconv"
	"time"
)

var (
	ipaddress        string
	toclean_etcdpath []string
	container_name   string
	external_domain  = "production"
)

type VulcandHostJSON struct {
	Host string ` json:"host"`
	Port int    `json:"port"`
}

func VulcandHostAdd(etcdnodes []string, dockeruri string, service fleet.SystemdService, port int, path string) error {

	// Docker internal ip address
	fmt.Println("Adding internal PROXY entry")
	if service.Id == 0 {
		container_name = service.Hostname
	} else {
		container_name = service.Hostname + "-" + strconv.FormatInt(service.Id, 10)
	}

	fullname := service.Hostname + "." + external_domain + "." + service.Region + "." + service.Domain

	rand.Seed(time.Now().UnixNano())

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(container_name)
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

	client, err := etcdbackend.NewEtcdBackend(registry.GetRegistry(), etcdnodes, "vulcand", "STRONG")
	if err != nil {
		return err
	}

	var host_exist backend.Host
	var endpointId, locationId, upstreamId string

	hosts, err := client.GetHosts()
	for _, host := range hosts {
		if host.Name == fullname {
			host_exist = *host
			break
		}
	}

	if host_exist.Name == "" {
		host, err := backend.NewHost(fullname)
		if err != nil {
			return err
		}
		host, err = client.AddHost(host)
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
		fmt.Println("Creating upstream")
		upstreamId = genId()
		upstream, err := backend.NewUpstream(upstreamId)
		if err != nil {
			return err
		}
		_, err = client.AddUpstream(upstream)
		if err != nil {
			return err
		}
	}

	if locationId == "" {
		fmt.Println("Creating location")
		locationId = strconv.FormatInt(service.Id, 10)
		location, err := backend.NewLocation(fullname, locationId, path, upstreamId)
		if err != nil {
			fmt.Println("Failed to create location")
			return err
		}
		_, err = client.AddLocation(location)
		if err != nil {
			fmt.Println("Failed to add location")
			return err
		}
	}

	if endpointId == "" {
		fmt.Println("Creating endpoint")
		endpointId = strconv.FormatInt(service.Id, 10)
		endpoint, err := backend.NewEndpoint(upstreamId, endpointId, "http://"+ipaddress+":"+strconv.Itoa(port))
		if err != nil {
			fmt.Println("Failed to create endpoint")
			return err
		}

		_, err = client.AddEndpoint(endpoint)
		if err != nil {
			fmt.Println("Failed to add endpoint")
			return err
		}
	}

	return nil
}

func VulcandHostDel(etcdnodes []string, dockeruri string, service fleet.SystemdService, port int, path string) error {

	fullname := service.Hostname + "." + external_domain + "." + service.Region + "." + service.Domain

	// Docker internal ip address
	fmt.Println("Remove internal PROXY entry")
	if service.Id == 0 {
		container_name = service.Hostname
	} else {
		container_name = service.Hostname + "-" + strconv.FormatInt(service.Id, 10)
	}

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(container_name)
		if err != nil {
			if i == 9 {
				fmt.Println("Failed to connect to docker")
				return err
			}
		} else {
			break
		}
		// wait 10 seconds to try again
		time.Sleep(10 * time.Second)
	}

	client, err := etcdbackend.NewEtcdBackend(registry.GetRegistry(), etcdnodes, "vulcand", "STRONG")
	if err != nil {
		return err
	}

	var endpointId, locationId, upstreamId string

	host, err := client.GetHost(fullname)
	if err != nil {
		return err
	}

	for _, location := range host.Locations {
		fmt.Println("Checking location path: '", location.Path, "'")
		if location.Path == path {
			locationId = location.GetId()
			upstreamId = location.Upstream.GetId()
			for _, endpoint := range location.Upstream.Endpoints {
				if endpoint.Url == "http://"+ipaddress+":"+strconv.Itoa(port) {
					endpointId = endpoint.GetId()
					break
					break
				}
			}
		}
	}

	if locationId != "" {
		fmt.Println("Deleting location")
		err = client.DeleteLocation(fullname, locationId)
		if err != nil {
			return err
		}
	}

	if endpointId != "" {
		fmt.Println("Deleting endpoint")
		err = client.DeleteEndpoint(upstreamId, endpointId)
		if err != nil {
			return err
		}
	}
	if upstreamId != "" {
		fmt.Println("Deleting upstream")
		err = client.DeleteUpstream(upstreamId)
		if err != nil {
			return err
		}
	}

	host, err = client.GetHost(fullname)
	if err != nil {
		return err
	}

	if len(host.Locations) == 0 {
		fmt.Println("Deleting host")
		err = client.DeleteHost(fullname)
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
