package dnslib

import (
	"github.com/DimShadoWWW/integrator/etcdlib"
	"strings"
)

func AddHostnameDNS(client *etcdlib.EtcdClient, id int64, hostname string, ipaddress string, port int, region string) error {

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
