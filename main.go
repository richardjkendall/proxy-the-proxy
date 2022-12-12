package main

import (
	"log"
	"net/http"
)

func main() {

	// Get my IP address
	myIpAddress := GetOutboundIP()
	log.Printf("My IP address is %s", myIpAddress)

	// get PAC content
	//pac := GetWpad()
	pac := `
	function FindProxyForURL(url, host) {
		return "DIRECT";
	}
	`

	handler := NewProxy(pac, myIpAddress.String())
	addr := "127.0.0.1:8080"
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalln("ListenAndServe", err)
	}

	// Run PAC code to get proxy
	//result := RunWpadPac(pac, myIpAddress.String())
	//log.Println(result)
}
