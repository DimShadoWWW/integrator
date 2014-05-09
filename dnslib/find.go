package dnslib

import (
	"encoding/json"
	"fmt"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"strings"
)

func GetHostnamesDNS(client *etcdlib.EtcdClient, path string) ([]DnsHost, error) {
	var paths []DnsHost
	recursive := true
	sorted := true
	r, err := client.Client.Get(path+"/", sorted, recursive)
	if err != nil {
		return nil, err
	}
	for _, a := range r.Node.Nodes {
		if a.Dir {
			for _, b := range a.Nodes {
				if b.Dir {
					for _, c := range b.Nodes {
						if c.Dir {
							for _, d := range c.Nodes {

								hostpath := strings.Split(strings.SplitAfterN(d.Key, "/", 3)[2], "/")
								for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
									hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
								}
								hostname := strings.Join(hostpath, ".")

								for _, p := range paths {

									if p.Hostname == hostname {

										var entry DnsEntry
										err := json.Unmarshal([]byte(d.Value), &entry)
										if err != nil {
											fmt.Println("error:", err)
											return nil, err
										}
										host := DnsHost{
											Hostname: hostname,
											EtcdKey:  d.Key,
											Entry:    []DnsEntry{entry},
										}
										paths = append(paths, host)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return paths, nil
}
