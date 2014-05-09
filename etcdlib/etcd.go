package etcdlib

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

type EtcdClient struct {
	client *etcd.Client
}

type DnsEntry struct {
	Port     int
	Priority int
	Host     string
}

type DnsHost struct {
	EtcdKey  string
	Hostname string
	Entry    DnsEntry
}

func (client *EtcdClient) GetDNSHostnames(path string) ([]DnsHost, error) {
	var paths []DnsHost
	recursive := true
	sorted := true
	r, err := client.client.Get(path+"/", sorted, recursive)
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
								var entry DnsEntry
								err := json.Unmarshal([]byte(d.Value), &entry)
								if err != nil {
									fmt.Println("error:", err)
									return nil, err
								}
								host := DnsHost{
									Hostname: hostname,
									EtcdKey:  d.Key,
									Entry:    entry,
								}
								paths = append(paths, host)
							}
						}
					}
				}
			}
		}
	}

	return paths, nil
}

func (client *EtcdClient) DeleteHostnameDNS(hostname, ipaddress string, port int) error {

	entry := DnsEntry{
		Host: ipaddress,
		Port: port,
	}
	hostpath := strings.Split(hostname, ".")
	for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
		hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
	}
	host := DnsHost{
		Hostname: hostname,
		EtcdKey:  "/skydns/" + strings.Join(hostpath, "/"),
		Entry:    entry,
	}

	if _, err := client.client.Delete(host.EtcdKey, true); err != nil {
		return formatErr(err)
	}

	hostpath = strings.Split(host.EtcdKey, "/")
	for i := 0; i < len(hostpath); i++ {
		local_path := "/" + strings.Join(hostpath[0:len(hostpath)-i], "/")
		response, err := client.client.Get(local_path, false, false)
		if notFound(err) {
			fmt.Printf("Key error: %s", err)
			// return nil
		} else {
			if isEmptyDir(response.Node) {
				client.client.Delete(response.Node.Key, true)
			} else {
				return nil
			}
		}
	}
	return nil
}

func (client *EtcdClient) CleanDNS() error {
	var paths []string
	recursive := true
	sorted := true
	r, err := client.client.Get("/skydns", sorted, recursive)
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
		response, err := client.client.Get(path, false, false)
		if notFound(err) {
			fmt.Printf("Key error: %s", err)
		}
		if isEmptyDir(response.Node) {
			client.client.Delete(response.Node.Key, true)
		} else {
			return nil
		}
	}
	return nil
}

func NewEtcdClient(nodes []string) (*EtcdClient, error) {
	client := etcd.NewClient(nodes)

	b := &EtcdClient{
		client: client,
	}
	return b, nil
}

func (s *EtcdClient) getVal(keys ...string) (string, bool) {
	response, err := s.client.Get(strings.Join(keys, "/"), false, false)
	if notFound(err) {
		return "", false
	}
	if isDir(response.Node) {
		return "", false
	}
	return response.Node.Value, true
}

func (s *EtcdClient) getDirs(keys ...string) []string {
	var out []string
	response, err := s.client.Get(strings.Join(keys, "/"), true, true)
	if notFound(err) {
		return out
	}

	if response == nil || !isDir(response.Node) {
		return out
	}

	for _, srvNode := range response.Node.Nodes {
		if isDir(srvNode) {
			out = append(out, srvNode.Key)
		}
	}
	return out
}

func (s *EtcdClient) getVals(keys ...string) []Pair {
	var out []Pair
	response, err := s.client.Get(strings.Join(keys, "/"), true, true)
	if notFound(err) {
		return out
	}

	if !isDir(response.Node) {
		return out
	}

	for _, srvNode := range response.Node.Nodes {
		if !isDir(srvNode) {
			out = append(out, Pair{srvNode.Key, srvNode.Value})
		}
	}
	return out
}

type Pair struct {
	Key string
	Val string
}

func suffix(key string) string {
	vals := strings.Split(key, "/")
	return vals[len(vals)-1]
}

func join(keys ...string) string {
	return strings.Join(keys, "/")
}

func formatErr(e error) error {
	switch err := e.(type) {
	case *etcd.EtcdError:
		return fmt.Errorf("Key error: %s", err.Message)
	}
	return e
}

func notFound(err error) bool {
	if err == nil {
		return false
	}
	eErr, ok := err.(*etcd.EtcdError)
	return ok && eErr.ErrorCode == 100
}

func isDupe(err error) bool {
	if err == nil {
		return false
	}
	eErr, ok := err.(*etcd.EtcdError)
	return ok && eErr.ErrorCode == 105
}

func isDir(n *etcd.Node) bool {
	return n != nil && n.Dir == true
}

func isEmptyDir(n *etcd.Node) bool {
	if isDir(n) && len(n.Nodes) == 0 {
		return true
	}
	return false
}
