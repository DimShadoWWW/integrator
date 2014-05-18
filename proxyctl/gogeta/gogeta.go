package gogeta

import (
	"encoding/json"
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"github.com/DimShadoWWW/integrator/fleet"
	// "github.com/fsouza/go-dockerclient"
	"strconv"
)

type GogetaHostJSON struct {
	Host string ` json:"host"`
	Port int    `json:"port"`
}

func GogetaHostAdd(client *etcdlib.EtcdClient, dockeruri string, service fleet.SystemdService, port int) error {

	dockerclient := dockerlib.NewDockerLib(dockeruri)

	ipaddress, err := dockerclient.GetContainerIpaddress(service.Hostname + "-" + strconv.FormatInt(service.Id, 10))
	if err != nil {
		return err
	}

	// without region
	_, err = client.Client.Create("/domains/"+service.Hostname+"."+service.Domain+"/type", "io", 0)
	if err != nil {
		return err
	}
	_, err = client.Client.Create("/domains/"+service.Hostname+"."+service.Domain+"/value", service.Hostname+"."+service.Domain, 0)
	if err != nil {
		return err
	}
	host_data, err := json.Marshal(GogetaHostJSON{
		Host: ipaddress,
		Port: port,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	client.Client.Create("/services/"+service.Hostname+"."+service.Domain+"/"+strconv.FormatInt(service.Id, 10)+"/location", string(host_data), 0)
	if err != nil {
		return err
	}

	// // with region
	// _, err = client.Client.Create("/domains/"+service.Region+"."+service.Hostname+"."+service.Domain+"/type", "io", 0)
	// if err != nil {
	// 	return err
	// }
	// _, err = client.Client.Create("/domains/"+service.Region+"."+service.Hostname+"."+service.Domain+"/value", service.Hostname+"."+service.Domain, 0)
	// if err != nil {
	// 	return err
	// }

	// host_data, err := json.Marshal(GogetaHostJSON{
	// 	Host: ipaddress,
	// 	Port: port,
	// })

	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	// client.Client.Create("/services/"+service.Region+"."+service.Hostname+"."+service.Domain+"/"+strconv.FormatInt(service.Id, 10)+"/location", string(host_data), 0)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func GogetaHostDel(client *etcdlib.EtcdClient, service fleet.SystemdService) error {

	_, err := client.Client.Delete("/services/"+service.Hostname+"."+service.Domain+"/"+strconv.FormatInt(service.Id, 10), true)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	response, err := client.Client.Get("/services/"+service.Hostname+"."+service.Domain, false, false)
	if etcdlib.NotFound(err) {
		fmt.Printf("Services not found: %s\n", err)
		return err

	} else {
		if etcdlib.IsEmptyDir(response.Node) {
			client.Client.Delete(response.Node.Key, false)
			_, err = client.Client.Delete("/domains/"+service.Hostname+"."+service.Domain, true)
			if err != nil {
				fmt.Printf("Can not delete domain: %s\n", err)
				return err
			}
			// _, err = client.Client.Delete("/domains/"+service.Region+"."+service.Hostname+"."+service.Domain, true)
			// if err != nil {
			// 	fmt.Printf("Can not delete domain: %s\n", err)
			// 	return err
			// }
		}

		return nil
	}

	return nil
}
