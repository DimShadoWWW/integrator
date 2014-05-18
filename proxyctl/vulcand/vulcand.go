package vulcand

import (
	"errors"
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type VulcandHostJSON struct {
	Host string ` json:"host"`
	Port int    `json:"port"`
}

func VulcandHostAdd(client *etcdlib.EtcdClient, dockeruri string, service fleet.SystemdService, port int, path string) error {
	rand.Seed(time.Now().UnixNano())

	fullname := service.Hostname + "." + service.Domain

	dockerclient := dockerlib.NewDockerLib(dockeruri)

	ipaddress, err := dockerclient.GetContainerIpaddress(service.Hostname + "-" + strconv.FormatInt(service.Id, 10))
	if err != nil {
		return err
	}

	// look for location..
	fmt.Println("Get: ", "/vulcand/hosts/"+fullname)
	response, err := client.Client.Get("/vulcand/hosts", false, false)
	if etcdlib.NotFound(err) {
		fmt.Printf("Host not found, creating..\n", err)
		fmt.Println("CreateDir: ", "/vulcand/hosts/")
		_, err = client.Client.CreateDir("/vulcand/hosts", 0)
		if err != nil {
			return err
		}
	}

	fmt.Println("Get: ", "/vulcand/hosts/"+fullname)
	response, err = client.Client.Get("/vulcand/hosts/"+fullname, false, false)
	if etcdlib.NotFound(err) {
		fmt.Printf("Host not found, creating..\n", err)
		fmt.Println("CreateDir: ", "/vulcand/hosts/"+fullname)
		_, err = client.Client.CreateDir("/vulcand/hosts/"+fullname, 0)
		if err != nil {
			return err
		}
	}

	// if location exist ..
	fmt.Println("Get: ", "/vulcand/hosts/"+fullname+"/locations")
	response, err = client.Client.Get("/vulcand/hosts/"+fullname+"/locations", false, false)
	if !etcdlib.NotFound(err) {
		for _, m := range response.Node.Nodes {
			fmt.Println(m.Key)

			fmt.Println("Get: ", m.Key+"/upstream")
			r, err := client.Client.Get(m.Key+"/upstream", false, false)
			if err != nil {
				return err
			}

			fmt.Println("Get: ", "/vulcand/upstreams/"+r.Node.Value+"/endpoints/")
			s, err := client.Client.Get("/vulcand/upstreams/"+r.Node.Value+"/endpoints/", false, false)
			if !etcdlib.NotFound(err) {
				for _, t := range s.Node.Nodes {
					fmt.Println("value: ", t.Value)
					fmt.Println("looking for: ", "http://"+ipaddress+":"+strconv.Itoa(port))
					if t.Value == "http://"+ipaddress+":"+strconv.Itoa(port) {
						return errors.New("Host already in endpoints")
					}
				}
			}
		}
	}

	fmt.Println("createUpstream")
	upstreamId, err := createUpstream(client)
	if err != nil {
		return err
	}

	fmt.Println("createEndpoint", string(upstreamId), ipaddress, strconv.Itoa(port))
	_, err = createEndpoint(client, upstreamId, ipaddress, port)
	if err != nil {
		return err
	}

	fmt.Println("createLocation", fullname, string(upstreamId), path)
	_, err = createLocation(client, fullname, upstreamId, path)
	if err != nil {
		return err
	}

	return nil
}

func VulcandHostDel(client *etcdlib.EtcdClient, dockeruri string, service fleet.SystemdService, port int) error {
	fullname := service.Hostname + "." + service.Domain

	dockerclient := dockerlib.NewDockerLib(dockeruri)

	ipaddress, err := dockerclient.GetContainerIpaddress(service.Hostname + "-" + strconv.FormatInt(service.Id, 10))
	if err != nil {
		return err
	}

	// if location exist ..
	fmt.Println("Get: ", "/vulcand/hosts/"+fullname+"/locations")
	response, err := client.Client.Get("/vulcand/hosts/"+fullname+"/locations", false, false)
	if !etcdlib.NotFound(err) {
		for _, m := range response.Node.Nodes {
			fmt.Println(m.Key)

			fmt.Println("Get: ", m.Key+"/upstream")
			r, err := client.Client.Get(m.Key+"/upstream", false, false)
			if err != nil {
				return err
			}

			fmt.Println("Get: ", "/vulcand/upstreams/"+r.Node.Value+"/endpoints/")
			s, err := client.Client.Get("/vulcand/upstreams/"+r.Node.Value+"/endpoints/", false, false)
			if !etcdlib.NotFound(err) {
				for _, t := range s.Node.Nodes {
					fmt.Println("value: ", t.Value)
					fmt.Println("looking for: ", "http://"+ipaddress+":"+strconv.Itoa(port))
					if t.Value == "http://"+ipaddress+":"+strconv.Itoa(port) {

						_, err := client.Client.Delete(m.Key, true)
						if err != nil {
							fmt.Printf("Key Error: %s\n", err)
							return err
						}

						_, err = client.Client.Delete(t.Key, true)
						if err != nil {
							fmt.Printf("Key Error: %s\n", err)
							return err
						}

						t1 := strings.Split(m.Key, "/")
						upstreamIdPath := strings.Join(t1[0:len(t1)-1], "/")
						upstId, err := client.Client.Get(upstreamIdPath, false, false)
						if err != nil {
							fmt.Printf("Key Error: %s\n", err)
						}
						if etcdlib.IsEmptyDir(upstId.Node) {

							_, err = client.Client.DeleteDir(upstreamIdPath)
							if err != nil {
								fmt.Printf("Key Error: %s\n", err)
							}

							t1 := strings.Split(upstreamIdPath, "/")
							upstreamPath := strings.Join(t1[0:len(t1)-1], "/")
							upstresp, err := client.Client.Get(upstreamPath, false, false)
							if err != nil {
								fmt.Printf("Key Error: %s\n", err)
							}
							if etcdlib.IsEmptyDir(upstresp.Node) {
								_, err = client.Client.DeleteDir(upstreamPath)
								if err != nil {
									fmt.Printf("Key Error: %s\n", err)
								}
							}
						}

						t2 := strings.Split(t.Key, "/")
						locationPath := strings.Join(t2[0:len(t2)-1], "/")
						locat, err := client.Client.Get(locationPath, false, false)
						if err != nil {
							fmt.Printf("Key Error: %s\n", err)
						}
						if etcdlib.IsEmptyDir(locat.Node) {

							_, err = client.Client.DeleteDir(locationPath)
							if err != nil {
								fmt.Printf("Key Error: %s\n", err)
							}

							t2 := strings.Split(locationPath, "/")
							hostPath := strings.Join(t2[0:len(t2)-1], "/")
							hostresp, err := client.Client.Get(hostPath, false, false)
							if err != nil {
								fmt.Printf("Key Error: %s\n", err)
							}
							if etcdlib.IsEmptyDir(hostresp.Node) {
								_, err = client.Client.DeleteDir(hostPath)
								if err != nil {
									fmt.Printf("Key Error: %s\n", err)
								}
							}
						}

						// echo etcdctl rm $i/path          # remove path
						// etcdctl rm $i/path          # remove path
						// echo etcdctl rm $i/upstream      # remove upstream id from host's locations
						// etcdctl rm $i/upstream      # remove upstream id from host's locations

						// echo etcdctl rmdir $i            # try to remove host's locations
						// etcdctl rmdir $i            # try to remove host's locations
						// echo etcdctl rmdir $j            # try to remove upstream's endpoints
						// etcdctl rmdir $j            # try to remove upstream's endpoints

						// echo etcdctl rmdir $(dirname $j) # remove upstream id
						// etcdctl rmdir $(dirname $j) # remove upstream id
						// echo etcdctl rmdir $(dirname $i) # remove host locations
						// etcdctl rmdir $(dirname $i) # remove host locations

						// echo etcdctl rmdir $(dirname $(dirname $j)) # remove upstream
						// etcdctl rmdir $(dirname $(dirname $j)) # remove upstream
						// echo etcdctl rmdir $(dirname $(dirname $i)) # remove host
						// etcdctl rmdir $(dirname $(dirname $i)) # remove host

						return nil
					}
				}
			}
		}
	}
	return errors.New("Endpoint not found")

	// _, err := client.Client.Delete("/services/"+service.Hostname+"."+service.Domain+"/"+strconv.FormatInt(service.Id, 10), true)
	// if err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// 	return err
	// }

	// response, err := client.Client.Get("/services/"+service.Hostname+"."+service.Domain, false, false)
	// if etcdlib.NotFound(err) {
	// 	fmt.Printf("Services not found: %s\n", err)
	// 	return err

	// } else {
	// 	if etcdlib.IsEmptyDir(response.Node) {
	// 		client.Client.Delete(response.Node.Key, false)
	// 		_, err = client.Client.Delete("/domains/"+service.Hostname+"."+service.Domain, true)
	// 		if err != nil {
	// 			fmt.Printf("Can not delete domain: %s\n", err)
	// 			return err
	// 		}
	// 	}

	// 	return nil
	// }

	// return nil
}

func createUpstream(client *etcdlib.EtcdClient) (string, error) {
	var id string
	for {
		id = genId()
		// create upstream
		fmt.Println("Get: ", "/vulcand/upstreams/"+id)
		_, err := client.Client.Get("/vulcand/upstreams/"+id, false, false)
		if etcdlib.NotFound(err) {
			// if not found it is available to be used
			break
		} // else try again
	}

	fmt.Println("CreateDir: ", "/vulcand/upstreams/"+id)
	_, err := client.Client.CreateDir("/vulcand/upstreams/"+id, 0)
	if err != nil {
		return "", err
	}
	return id, nil
}

func createEndpoint(client *etcdlib.EtcdClient, upstreamId string, ipaddress string, port int) (string, error) {
	var id string
	for {
		id = genId()
		// create upstream
		fmt.Println("Get: ", "/vulcand/upstreams/"+upstreamId+"/endpoints/"+id)
		_, err := client.Client.Get("/vulcand/upstreams/"+upstreamId+"/endpoints/"+id, false, false)
		if etcdlib.NotFound(err) {
			// if not found it is available to be used
			break
		} // else try again
	}

	fmt.Println("Set: ", "/vulcand/upstreams/"+upstreamId+"/endpoints/"+id, "http://"+ipaddress+":"+strconv.Itoa(port))
	_, err := client.Client.Set("/vulcand/upstreams/"+upstreamId+"/endpoints/"+id, "http://"+ipaddress+":"+strconv.Itoa(port), 0)
	if err != nil {
		return "", err
	}
	return id, nil
}

func createLocation(client *etcdlib.EtcdClient, fullname string, upstreamId string, path string) (string, error) {
	var id string
	for {
		id = genId()
		// create upstream
		fmt.Println("Get: ", "/vulcand/hosts/"+fullname+"/locations/"+id)
		_, err := client.Client.Get("/vulcand/hosts/"+fullname+"/locations/"+id, false, false)
		if etcdlib.NotFound(err) {
			// if not found it is available to be used
			break
		} // else try again
	}

	fmt.Println("Set: ", "/vulcand/hosts/"+fullname+"/locations/"+id+"/path", path)
	_, err := client.Client.Set("/vulcand/hosts/"+fullname+"/locations/"+id+"/path", path, 0)
	if err != nil {
		return "", err
	}

	fmt.Println("Set: ", "/vulcand/hosts/"+fullname+"/locations/"+id+"/upstream", upstreamId)
	_, err = client.Client.Set("/vulcand/hosts/"+fullname+"/locations/"+id+"/upstream", upstreamId, 0)
	if err != nil {
		return "", err
	}
	return id, nil
}

func genId() string {
	var id int64
	for i := 0; i < 10; i++ {
		id = rand.Int63() + 1
	}
	return strconv.FormatInt(id, 10)
}
