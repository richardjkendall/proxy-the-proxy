package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		log.Fatalln()
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	log.Printf("GetOutboundIP: IP = %v", localAddr)
	return localAddr.IP
}

func CountConsecutiveCharacters(src string, tgt string) int {
	count := 0
	for _, c := range src {
		if string(c) == tgt {
			count++
		} else {
			break
		}
	}
	return count
}

func CountBitsInMask(mask string) (int, error) {
	octets := strings.Split(mask, ".")
	count := 0
	binarymask := ""
	for _, octet := range octets {
		octet_int, err := strconv.Atoi(octet)
		if err != nil {
			return count, err
		}
		binarymask = fmt.Sprintf("%s%b", binarymask, octet_int)
	}
	count = CountConsecutiveCharacters(binarymask, "1")
	log.Printf("CountBitsInMask: Mask = %v, Binary = %v, Count of 1s = %v", mask, binarymask, count)
	return count, nil
}

func GetNetFromIPandMask(ip string, mask string) (*net.IPNet, error) {
	masksize, err := CountBitsInMask(mask)
	if err != nil {
		return nil, err
	}
	cidr := fmt.Sprintf("%s/%d", ip, masksize)
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	log.Printf("GetNetFromIPandMask: Ip = %v, Mask = %v, Net = %v", ip, mask, ipNet)
	return ipNet, nil
}

func IsIpInRange(ip string, pattern string, mask string) (bool, error) {
	IPAddress := net.ParseIP(ip)
	IPNet, err := GetNetFromIPandMask(pattern, mask)
	if err != nil {
		return false, err
	}
	log.Printf("IsIpInRange: IP = %v, IPnet = %v", IPAddress, IPNet)
	return IPNet.Contains(IPAddress), nil
}

func IpToDecimal(ip string) int {
	ipaddr := net.ParseIP(ip).To4()
	log.Printf("IpToDecimal: IP = %v", ipaddr)
	if ipaddr != nil {
		shifted := int(ipaddr[0])<<24 |
			int(ipaddr[1])<<16 |
			int(ipaddr[2])<<8 |
			int(ipaddr[3])
		log.Printf("IpToDecimal: decimal version = %v", shifted)
		return shifted
	} else {
		log.Printf("IpToDecimal: invalid IP")
		return 0
	}
}
