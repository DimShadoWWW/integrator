package vulcand

import (
	"errors"
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/mailgun/vulcand/backend"
	"github.com/mailgun/vulcand/backend/etcdbackend"
	"github.com/mailgun/vulcand/plugin/registry"
	"math/rand"
	"strconv"
	"time"
	// "strings"
	// "github.com/DimShadoWWW/integrator/etcdlib"
	// "github.com/coreos/go-etcd/etcd"
	// vulcanserv "github.com/mailgun/vulcand/service"
)

var (
	ipaddress        string
	toclean_etcdpath []string
	external_domain  string
	container_name   string
)

type VulcandHostJSON struct {
	Host string ` json:"host"`
	Port int    `json:"port"`
}

func VulcandHostAdd(etcdnodes []string, dockeruri string, service fleet.SystemdService, port int, path string) error {

	external_domain = "production" // region used for

	container_name = service.Hostname + "-" + strconv.FormatInt(service.Id, 10)
	fullname := service.Hostname + "." + external_domain + "." + service.Region + "." + service.Domain
	// service.Hostname + "." + service.Domain

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

	// options := vulcanserv.Options{}
	// options.EtcdNodes = etcdnodes
	// serv := vulcanserv.NewService(options, registry.GetRegistry())

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
			fmt.Println("Hosts exist")
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
		fmt.Println("createUpstream")
		upstreamId = genId()
		upstream, err := backend.NewUpstream(upstreamId)
		if err != nil {
			return err
		}
		upstream, err = client.AddUpstream(upstream)
		if err != nil {
			return err
		}
	}

	if locationId == "" {
		fmt.Println("createLocation")
		locationId = strconv.FormatInt(service.Id, 10)
		location, err := backend.NewLocation(locationId, fullname, path, upstreamId)
		if err != nil {
			return err
		}
		location, err = client.AddLocation(location)
		if err != nil {
			return err
		}
	}

	if endpointId == "" {
		fmt.Println("createEndpoint")
		endpointId = strconv.FormatInt(service.Id, 10)
		endpoint, err := backend.NewEndpoint(upstreamId, endpointId, "http://"+ipaddress+":"+strconv.Itoa(port))
		if err != nil {
			return err
		}

		endpoint, err = client.AddEndpoint(endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func VulcandHostDel(etcdnodes []string, dockeruri string, service fleet.SystemdService, port int, path string) error {

	fullname := service.Hostname + "." + service.Domain
	container_name = service.Hostname + "-" + strconv.FormatInt(service.Id, 10)

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

	var endpointId, locationId, upstreamId string

	// var host_exist backend.Host
	// hosts, err := client.GetHosts()
	// for _, host := range hosts {
	// 	if host.Name == fullname {
	// 		host_exist = *host
	// 		fmt.Println("Hosts exist")
	// 		break
	// 	}
	// }

	// if host_exist.Name == "" {
	// 	host, err := backend.NewHost(fullname)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	host, err = client.AddHost(host)
	// 	if err != nil {
	// 		return err
	// 	}
	// } else {
	// 	fmt.Println("Checking hosts")
	// 	for _, location := range host_exist.Locations {
	// 		fmt.Println("Checking location", location.Hostname)
	// 		if location.Path == path {
	// 			locationId = location.Id
	// 			upstreamId = location.Upstream.Id
	// 			for _, endpoint := range location.Upstream.Endpoints {
	// 				if endpoint.Url == "http://"+ipaddress+":"+strconv.Itoa(port) {
	// 					toclean_etcdpath = append(toclean_etcdpath, location.EtcdKey)
	// 					toclean_etcdpath = append(toclean_etcdpath, location.Upstream.EtcdKey)
	// 					toclean_etcdpath = append(toclean_etcdpath, endpoint.EtcdKey)
	// 					endpointId = endpoint.Id
	// 					fmt.Println(endpoint.Stats)
	// 					break
	// 					break
	// 				}
	// 			}
	// 		}
	// 	}
	// }

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

	// etcdclient := etcd.NewClient(etcdnodes)

	// for _, key := range toclean_etcdpath {
	// 	key_sp := strings.Split(key, "/")
	// 	c := len(key_sp)
	// 	for i := 0; i < c; i++ {
	// 		k := strings.Join(key_sp[0:c-i], "/")
	// 		fmt.Println(k)
	// 		req, err := etcdclient.Get(k, false, false)
	// 		if err == nil {
	// 			if req.Node.Dir {
	// 				if etcdlib.IsEmptyDir(req.Node) {
	// 					etcdclient.DeleteDir(req.Node.Key)
	// 				}
	// 			}
	// 		}

	// 	}

	// }

	return nil
}

func genId() string {
	var id int64
	for i := 0; i < 10; i++ {
		id = rand.Int63() + 1
	}
	return string(strconv.FormatInt(id, 10))
}
