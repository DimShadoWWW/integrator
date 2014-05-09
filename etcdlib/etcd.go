package etcdlib

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

type EtcdClient struct {
	Client *etcd.Client
}

func NewEtcdClient(nodes []string) (*EtcdClient, error) {
	client := etcd.NewClient(nodes)

	b := &EtcdClient{
		Client: client,
	}
	return b, nil
}

func (s *EtcdClient) getVal(keys ...string) (string, bool) {
	response, err := s.Client.Get(strings.Join(keys, "/"), false, false)
	if NotFound(err) {
		return "", false
	}
	if IsDir(response.Node) {
		return "", false
	}
	return response.Node.Value, true
}

func (s *EtcdClient) getDirs(keys ...string) []string {
	var out []string
	response, err := s.Client.Get(strings.Join(keys, "/"), true, true)
	if NotFound(err) {
		return out
	}

	if response == nil || !IsDir(response.Node) {
		return out
	}

	for _, srvNode := range response.Node.Nodes {
		if IsDir(srvNode) {
			out = append(out, srvNode.Key)
		}
	}
	return out
}

func (s *EtcdClient) getVals(keys ...string) []Pair {
	var out []Pair
	response, err := s.Client.Get(strings.Join(keys, "/"), true, true)
	if NotFound(err) {
		return out
	}

	if !IsDir(response.Node) {
		return out
	}

	for _, srvNode := range response.Node.Nodes {
		if !IsDir(srvNode) {
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

func FormatErr(e error) error {
	switch err := e.(type) {
	case *etcd.EtcdError:
		return fmt.Errorf("Key error: %s", err.Message)
	}
	return e
}

func NotFound(err error) bool {
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

func IsDir(n *etcd.Node) bool {
	return n != nil && n.Dir == true
}

func IsEmptyDir(n *etcd.Node) bool {
	if IsDir(n) && len(n.Nodes) == 0 {
		return true
	}
	return false
}
