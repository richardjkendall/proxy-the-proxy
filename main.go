package main

import "log"

func main() {

	// Get my IP address
	myIpAddress := GetOutboundIP()
	log.Printf("My IP address is %s", myIpAddress)

	// get PAC content
	pac := GetWpad()

	// Run PAC code to get proxy
	result := RunWpadPac(pac, myIpAddress.String())
	log.Println(result)
}
