package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/robertkrimen/otto"
)

func GetWpad() string {
	// get PAC file
	resp, err := http.Get("http://127.0.0.1/wpad.dat")
	if err != nil {
		log.Fatalln(err)
	}

	// read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)
	return sb
}

func RunWpadPac(pac string, ipaddress string) string {
	vm := otto.New()

	vm.Set("myIpAddress", func(call otto.FunctionCall) otto.Value {
		result, _ := vm.ToValue(ipaddress)
		return result
	})

	vm.Set("dnsDomainIs", func(call otto.FunctionCall) otto.Value {
		host, _ := call.Argument(0).ToString()
		domain, _ := call.Argument(1).ToString()
		log.Printf(`dnsDomainIs: host %v, domain %v`, host, domain)
		if host == "undefined" || domain == "undefined" {
			log.Printf(`dnsDomainIs: one or both of the parameters are undefined`)
			result, _ := vm.ToValue(false)
			return result
		}
		result, _ := vm.ToValue(strings.HasSuffix(host, domain))
		return result
	})

	vm.Set("localHostOrDomainIs", func(call otto.FunctionCall) otto.Value {
		host, _ := call.Argument(0).ToString()
		hostdom, _ := call.Argument(1).ToString()
		log.Printf(`localHostOrDomainIs: host %v, hostdom %v`, host, hostdom)
		if host == "undefined" || hostdom == "undefined" {
			log.Printf(`localHostOrDomainIs: one or both of the parameters are undefined`)
			result, _ := vm.ToValue(false)
			return result
		}
		dommatch := strings.HasSuffix(host, hostdom)
		hostmatch := false
		hostdombits := strings.Split(hostdom, ".")
		if len(hostdombits) > 0 {
			hostmatch = host == hostdombits[0]
		}
		result, _ := vm.ToValue(dommatch || hostmatch)
		return result
	})

	vm.Set("isPlainHostName", func(call otto.FunctionCall) otto.Value {
		host, _ := call.Argument(0).ToString()
		if host == "undefined" {
			log.Printf(`isPlainHostName: host is undefined`)
			result, _ := vm.ToValue(false)
			return result
		}
		result, _ := vm.ToValue(!strings.Contains(host, "."))
		return result
	})

	vm.Set("shExpMatch", func(call otto.FunctionCall) otto.Value {
		str, _ := call.Argument(0).ToString()
		shexp, _ := call.Argument(1).ToString()
		match, err := filepath.Match(shexp, str)
		if err != nil {
			log.Printf("shExpMatch: error doing match")
		}
		result, _ := vm.ToValue(match)
		return result
	})

	vm.Set("dnsResolve", func(call otto.FunctionCall) otto.Value {
		dns, _ := call.Argument(0).ToString()
		result, _ := vm.ToValue(PerformDNSLookup(dns))
		return result
	})

	vm.Set("isResolvable", func(call otto.FunctionCall) otto.Value {
		dns, _ := call.Argument(0).ToString()
		lookup := PerformDNSLookup(dns)
		switch lookup.(type) {
		case bool:
			result, _ := vm.ToValue(lookup)
			return result
		default:
			result, _ := vm.ToValue(true)
			return result
		}
	})

	vm.Set("isInNet", func(call otto.FunctionCall) otto.Value {
		host, _ := call.Argument(0).ToString()
		pattern, _ := call.Argument(1).ToString()
		mask, _ := call.Argument(2).ToString()

		inrange, err := IsIpInRange(host, pattern, mask)
		if err != nil {
			result, _ := vm.ToValue(false)
			return result
		}
		result, _ := vm.ToValue(inrange)
		return result
	})

	vm.Set("convert_addr", func(call otto.FunctionCall) otto.Value {
		ip, _ := call.Argument(0).ToString()
		ipdecimal := IpToDecimal(ip)
		result, _ := vm.ToValue(ipdecimal)
		return result
	})

	vm.Set("dnsDomainLevels", func(call otto.FunctionCall) otto.Value {
		dns, _ := call.Argument(0).ToString()
		levels := len(strings.Split(dns, ".")) - 1
		result, _ := vm.ToValue(levels)
		return result
	})

	_, err := vm.Run(fmt.Sprintf(`
	%s

	var output = FindProxyForURL("https://www.google.com", "www.google.com");

	`, pac))

	if err != nil {
		log.Fatalln(err)
	}

	// need to get output value
	output, err := vm.Get("output")
	if err != nil {
		log.Fatalln(err)
	}
	output_string, err := output.ToString()
	if err != nil {
		log.Fatalln(err)
	}

	return output_string
}
