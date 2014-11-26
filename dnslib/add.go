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
	out_ipaddress   string
	out_port        int
)

func AddHostnameDNS(client *etcdlib.EtcdClient, dockeruri string, id int64, name string, hostname string, domain string, port int, protocol string, region string, priority int, iface string) error {

	external_domain = "production" // region used for
	internal_domain = "docker"     // region used for
	server_hostname := region
	// server_hostname, err := os.Hostname()
	// if err != nil {
	// 	return err
	// }

	out_hostname = strconv.FormatInt(id, 10) + "." + hostname + "." + external_domain + "." + region + "." + domain
	in_hostname = strconv.FormatInt(id, 10) + "." + hostname + "." + internal_domain + "." + server_hostname + "." + domain

	fmt.Println("internal hostname", in_hostname)

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	// Docker internal ip address
	fmt.Println("Adding internal DNS entry")
	container_name = name

	retries := 31
	for i := 0; i < retries; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(container_name)
		if err != nil {
			// retry until this
			if i == retries-1 {
				fmt.Println("Failed to get docker's container ip address")
				return err
			}
		} else {
			break
		}

		// wait 10 seconds to try again
		time.Sleep(10 * time.Second)
	}
	// Two seconds waiting for stabilize the docker container
	time.Sleep(2 * time.Second)

	if protocol == "tcp" {
		p, err := dockerclient.GetContainerTcpPort(container_name, port)
		if err != nil {
			fmt.Println(err)
			out_port = 0
		}
		fmt.Println("Tcp port ", port, " listening on ", p)
		out_port, err = strconv.Atoi(p)
		if err != nil {
			fmt.Println(err)
			out_port = 0
		}
	} else {
		p, err := dockerclient.GetContainerUdpPort(container_name, port)
		if err != nil {
			fmt.Println(err)
			out_port = 0
		}
		fmt.Println("Udp port ", port, " listening on ", p)
		out_port, err = strconv.Atoi(p)
		if err != nil {
			fmt.Println(err)
			out_port = 0
		}
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
	if _, err = client.Client.Set(host.EtcdKey, string(json_data), 0); err != nil {
		return err
	}

	// Add CoreOS's node external ip address to hostname with region
	fmt.Println("Adding external DNS entry")

	if iface == "" {
		out_ipaddress, err = GetLocalIp("8.8.8.8:53")
		if err != nil {
			return err
		}
	} else {
		out_ipaddress, err = GetLocalIpByInterface(iface)
		if err != nil {
			out_ipaddress, err = GetLocalIp("8.8.8.8:53")
			if err != nil {
				return err
			}
		}
	}

	entry = DnsEntry{
		Host:     out_ipaddress,
		Port:     out_port,
		Priority: priority,
	}
	fmt.Printf("DnsEntry: %#v\n", entry)
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

	if _, err = client.Client.CreateDir(strings.Join(hostpath[0:len(hostpath)-1], "/"), 20); err != nil {
		return err
	}

	fmt.Println("Add ", host.EtcdKey)
	if _, err = client.Client.Set(host.EtcdKey, string(json_data), 0); err != nil {
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
