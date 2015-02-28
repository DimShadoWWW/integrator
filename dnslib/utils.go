package dnslib

import (
	"net"
	"strings"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func GetLocalIp(server string) (string, error) {
	conn, err := net.Dial("udp", server)
	if err != nil {
		return "", err
	}

	// conn.LocalAddr().String() returns ip_address:port
	return net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0]).String(), nil
}

func GetLocalIpByInterface(interf string) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range ifaces {
		if i.Name == interf {
			addrs, err := i.Addrs()
			if err != nil {
				return "", err
			}
			return addrs[0].(*net.IPNet).IP.String(), nil
			// for _, a := range addrs {
			// 	switch v := a.(type) {
			// 	case *net.IPAddr:
			// 		fmt.Printf("%v : %s (%s)\n", i.Name, v, v.IP.DefaultMask())
			// 	}

			// }
		}
	}
	return "", Error("Network interface not found")
}
