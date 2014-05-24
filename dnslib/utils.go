package dnslib

import (
	"net"
	"strings"
)

func GetLocalIp(server string) (string, error) {
	conn, err := net.Dial("udp", server)
	if err != nil {
		return "", err
	}

	// conn.LocalAddr().String() returns ip_address:port
	return net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0]).String(), nil
}
