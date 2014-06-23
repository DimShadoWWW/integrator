package dnslib

import (
	"encoding/json"
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"strconv"
	"strings"
	"time"
)

var (
	ipaddress       string
	container_name  string
	out_hostname    string
	in_hostname     string
	internal_domain string
	external_domain string
	out_port        string
)

func AddHostnameDNS(client *etcdlib.EtcdClient, dockeruri string, id int64, hostname string, domain string, port int, region string, priority int) error {

	external_domain = "production" // region used for
	internal_domain = "docker"     // region used for

	out_hostname = strconv.FormatInt(id, 10) + "." + hostname + "." + external_domain + "." + region + "." + domain
	in_hostname = strconv.FormatInt(id, 10) + "." + hostname + "." + internal_domain + "." + region + "." + domain

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	// Docker internal ip address
	fmt.Println("Adding internal DNS entry")
	if id == 0 {
		container_name = hostname
	} else {
		container_name = hostname + "-" + strconv.FormatInt(id, 10)
	}

	for i := 0; i < 10; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(container_name)
		if err != nil {
			fmt.Println(err)
			// retry until this
			if i == 9 {
				return err
			}
		} else {
			break
		}
		// wait 10 seconds to try again
		time.Sleep(10 * time.Second)
	}

	fmt.Println("Creating entry")

	entry := DnsEntry{
		Host:     ipaddress,
		Port:     port,
		Priority: priority,
	}

	hostpath := strings.Split(in_hostname, ".")
	for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
		hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
	}

	host := DnsHost{
		Hostname: in_hostname,
		EtcdKey:  "/skydns/" + strings.Join(hostpath, "/"),
		Entry:    []DnsEntry{entry},
	}

	json_data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	fmt.Println("Add ", host.EtcdKey)
	if _, err = client.Client.Set(host.EtcdKey, string(json_data), 20); err != nil {
		return err
	}

	// Add CoreOS's node external ip address to hostname with region
	fmt.Println("Adding external DNS entry")
	out_ipaddress, err := GetLocalIp("8.8.8.8:53")
	if err != nil {
		return err
	}

	if port == 0 {
		out_port = "0"
	} else {
		for i := 0; i < 10; i++ {
			out_port, err = dockerclient.GetContainerTcpPort(container_name, port)
			if err != nil {
				// retry until this
				fmt.Println(err)
				if i == 9 {
					return err
				}
			} else {
				break
			}
			// wait 10 seconds to try again
			time.Sleep(10 * time.Second)
		}
	}

	oport, err := strconv.Atoi(out_port)
	if err != nil {
		return err
	}

	entry = DnsEntry{
		Host:     out_ipaddress,
		Port:     oport,
		Priority: priority,
	}
	json_data, err = json.Marshal(entry)
	if err != nil {
		return err
	}

	hostpath = strings.Split(out_hostname, ".")
	for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
		hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
	}

	host = DnsHost{
		Hostname: out_hostname,
		EtcdKey:  "/skydns/" + strings.Join(hostpath, "/"),
		Entry:    []DnsEntry{entry},
	}

	fmt.Println("Add ", host.EtcdKey)
	if _, err = client.Client.Set(host.EtcdKey, string(json_data), 20); err != nil {
		return err
	}
	return nil
}

func FindHostnameByIP(client *etcdlib.EtcdClient, address string) string {
	hostnames, err := GetHostnamesDNS(client, "/skydns/")
	if err != nil {
		return ""
	}
	for _, host := range hostnames {
		for _, entry := range host.Entry {
			if entry.Host == address {
				hostpath := strings.Split(strings.SplitAfterN(host.EtcdKey, "/", 3)[2], "/")
				for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
					hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
				}
				hostname := strings.Join(hostpath, ".")
				return hostname
			}
		}
	}
	return ""
}
