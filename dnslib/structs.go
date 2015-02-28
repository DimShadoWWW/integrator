package dnslib

type DnsEntry struct {
	Port     int
	Priority int
	Host     string
}

type DnsHost struct {
	EtcdKey  string
	Hostname string
	Entry    []DnsEntry
}
