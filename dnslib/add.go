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

func AddHostnameDNS(client *etcdlib.EtcdClient, dockeruri string, id int64, hostname string, domain string, port int, region string, priority int) error {

	internal_domain := "docker" // region used for

	out_hostname := region + "." + hostname + "." + domain
	in_hostname := region + "." + hostname + "." + internal_domain + "." + domain

	dockerclient := dockerlib.NewDockerLib(dockeruri)

	// Docker internal ip address
	var ipaddress string
	var err error

	for i := 0; i < 10; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(hostname + "-" + strconv.FormatInt(id, 10))
		if err != nil {
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

	entry := DnsEntry{
		Host:     ipaddress,
		Port:     port,
		Priority: priority,
	}

	hostpath := strings.Split(strconv.FormatInt(id, 10)+"."+in_hostname, ".")
	for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
		hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
	}

	host := DnsHost{
		Hostname: strconv.FormatInt(id, 10) + "." + in_hostname,
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
	out_ipaddress, err := getLocalIp("8.8.8.8:53")
	if err != nil {
		return err
	}
	var out_port string

	for i := 0; i < 10; i++ {
		out_port, err = dockerclient.GetContainerTcpPort(hostname+"-"+strconv.FormatInt(id, 10), port)
		if err != nil {
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

	hostpath = strings.Split(strconv.FormatInt(id, 10)+"."+out_hostname, ".")
	for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
		hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
	}

	host = DnsHost{
		Hostname: strconv.FormatInt(id, 10) + "." + out_hostname,
		EtcdKey:  "/skydns/" + strings.Join(hostpath, "/"),
		Entry:    []DnsEntry{entry},
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
