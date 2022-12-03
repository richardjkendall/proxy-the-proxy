package main

import (
	"log"
	"net"
)

func PerformDNSLookup(host string) interface{} {
	log.Printf("PerformDNSLookup: %s", host)
	ips, err := net.LookupIP(host)
	if err != nil {
		log.Println(err)
		return false
	}
	if len(ips) > 0 {
		return ips[0].String()
	} else {
		return false
	}
}
