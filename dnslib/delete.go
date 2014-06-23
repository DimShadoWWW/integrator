package dnslib

import (
	"fmt"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"strconv"
	"strings"
)

func DeleteHostnameDNS(client *etcdlib.EtcdClient, id int64, hostname string, domain string, port int, region string) error {

	external_domain = "production" // region used for
	internal_domain = "docker"     // region used for
	out_hostname = strconv.FormatInt(id, 10) + "." + hostname + "." + external_domain + "." + region + "." + domain
	in_hostname = strconv.FormatInt(id, 10) + "." + hostname + "." + internal_domain + "." + region + "." + domain

	// Docker internal ip address

	hostpath := strings.Split(in_hostname, ".")
	for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
		hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
	}

	host := DnsHost{
		Hostname: in_hostname,
		EtcdKey:  "/skydns/" + strings.Join(hostpath, "/"),
		Entry:    []DnsEntry{},
	}

	fmt.Println("Remove ", host.EtcdKey)
	if _, err := client.Client.Delete(host.EtcdKey, true); err != nil {
		fmt.Println(err)
	}

	hostpath = strings.Split(host.EtcdKey, "/")
	for i := 1; i < len(hostpath); i++ {
		local_path := "/" + strings.Join(hostpath[0:len(hostpath)-i], "/")
		response, err := client.Client.Get(local_path, false, false)
		if etcdlib.NotFound(err) {
			fmt.Printf("Key error: %s", err)
			// return nil
		} else {
			if etcdlib.IsEmptyDir(response.Node) {
				fmt.Println("Remove ", response.Node.Key)
				client.Client.Delete(response.Node.Key, true)
			}
		}
	}

	hostpath = strings.Split(out_hostname, ".")
	for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
		hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
	}

	host = DnsHost{
		Hostname: out_hostname,
		EtcdKey:  "/skydns/" + strings.Join(hostpath, "/"),
		Entry:    []DnsEntry{},
	}

	fmt.Println("Remove ", host.EtcdKey)
	if _, err := client.Client.Delete(host.EtcdKey, true); err != nil {
		return fmt.Errorf("Host not found: %s", hostname)
	}

	hostpath = strings.Split(host.EtcdKey, "/")
	for i := 1; i < len(hostpath); i++ {
		local_path := "/" + strings.Join(hostpath[0:len(hostpath)-i], "/")
		response, err := client.Client.Get(local_path, false, false)
		if etcdlib.NotFound(err) {
			fmt.Printf("Key error: %s", err)
			// return nil
		} else {
			if etcdlib.IsEmptyDir(response.Node) {
				fmt.Println("Remove ", response.Node.Key)
				client.Client.Delete(response.Node.Key, true)
			}
		}
	}
	return nil
}

func CleanDNS(client *etcdlib.EtcdClient) error {
	var paths []string
	recursive := true
	sorted := true
	r, err := client.Client.Get("/skydns", sorted, recursive)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, a := range r.Node.Nodes {
		if a.Dir {
			for _, b := range a.Nodes {
				if b.Dir {
					for _, c := range b.Nodes {
						if c.Dir {
							paths = append([]string{c.Key}, paths...)
						}
					}
					paths = append(paths, b.Key)
				}
			}
			paths = append(paths, a.Key)
		}
	}
	for _, path := range paths {
		response, err := client.Client.Get(path, false, false)
		if etcdlib.NotFound(err) {
			fmt.Printf("Key error: %s", err)
		}
		if etcdlib.IsEmptyDir(response.Node) {
			client.Client.Delete(response.Node.Key, true)
		} else {
			return nil
		}
	}
	return nil
}
