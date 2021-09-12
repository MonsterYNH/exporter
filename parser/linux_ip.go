package parser

import (
	"errors"
	"net"
)

func (parser *LinuxParser) ParseIPStat() (IPStat, error) {
	return GetIPs()
}

func GetIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	if len(ips) == 0 {
		return nil, errors.New("not found ipv4")
	}

	return ips, nil
}
