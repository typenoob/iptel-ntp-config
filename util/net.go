package util

import "net"

func IPToInt(ip net.IP) int {
	ip = ip.To4()
	return int(ip[0])<<24 | int(ip[1])<<16 | int(ip[2])<<8 | int(ip[3])
}

func IntToIP(n int) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

// 遍历 CIDR 范围内的所有 IP 地址
func IterateCIDR(ipNet *net.IPNet, handler func(net.IP)) {
	mask := ipNet.Mask
	network := ipNet.IP.Mask(mask)
	broadcast := make(net.IP, len(network))
	for i := range network {
		broadcast[i] = network[i] | ^mask[i]
	}
	start := IPToInt(network)
	end := IPToInt(broadcast)

	for i := start + 1; i < end; i++ {
		ip := IntToIP(i)
		handler(ip)
	}
}
