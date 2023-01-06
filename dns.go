package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"strings"
)

type ResolveConf struct {
	search     [][]string
	nameserver []string
}

func GetSearchDomain() []string {
	domain, err := GetDomainFromHostname()
	if err != nil {
		log.Printf(`GetSearchDomain: error getting domain from hostname: %v`, err)
		resolvconf, err := ReadResolvConf()
		if err != nil {
			log.Printf(`GetSearchDomain: error calling ReadResolvConf: %v`, err)
		}
		log.Printf(`GetSearchDomain: response from ReadResolvConf = %v`, resolvconf)
		if len(resolvconf.search) > 0 {
			if len(resolvconf.search[0]) > 0 {
				domains := []string{}
				for _, doms := range resolvconf.search {
					for _, domain := range doms {
						domains = append(domains, domain)
					}
				}
				log.Printf(`GetSearchDomain: flattened list of search domains = %v`, domains)
				return domains
			}
		}
		log.Printf(`GetSearchDomain: returning blank domain`)
		return []string{}
	}
	log.Printf(`GetSearchDomain: returning domain from hostname = %s`, domain)
	return []string{domain}
}

func ReadResolvConf() (ResolveConf, error) {
	// based on this https://man7.org/linux/man-pages/man5/resolv.conf.5.html
	rc := ResolveConf{}
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		log.Printf(`ReadResolvConf: error opening /etc/resolv.conf = %v`, err)
		return rc, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			// this is a nameserver line
			fields := strings.Fields(line)
			if len(fields) > 1 {
				rc.nameserver = append(rc.nameserver, fields[len(fields)-1])
			} else {
				log.Printf(`ReadResolvConf: malformed nameserver line = %s`, line)
			}
		}
		if strings.HasPrefix(line, "search") {
			// this is a search domain line
			fields := strings.Fields(line)
			if len(fields) > 1 {
				rc.search = append(rc.search, fields[1:])
			} else {
				log.Printf(`ReadResolvConf: malformed search line = %s`, line)
			}
		}
	}
	file.Close()
	return rc, nil
}

func GetDomainFromHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf(`GetDomainFromHostname: error getting hostname = %v`, hostname)
		return "", err
	}
	hostbits := strings.Split(hostname, ".")
	if len(hostbits) > 1 {
		return strings.Join(hostbits[1:], "."), nil
	} else {
		log.Printf(`GetDomainFromHostname: no domain name available in host name`)
		return hostname, errors.New("No domain available in host name")
	}
}

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
