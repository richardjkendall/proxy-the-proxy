package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const motd string = `
   ___                       _   _                                      
  / _ \_ __ _____  ___   _  | |_| |__   ___   _ __  _ __ _____  ___   _ 
 / /_)/ '__/ _ \ \/ / | | | | __| '_ \ / _ \ | '_ \| '__/ _ \ \/ / | | |
/ ___/| | | (_) >  <| |_| | | |_| | | |  __/ | |_) | | | (_) >  <| |_| |
\/    |_|  \___/_/\_\\__, |  \__|_| |_|\___| | .__/|_|  \___/_/\_\\__, |
                     |___/                   |_|                  |___/ 

By Richard Kendall
Released under MIT licence
https://github.com/richardjkendall/proxy-the-proxy
v%v

`

var global_proxy *proxy

func CreateMgmtServer(port int) *http.Server {

	type resp struct {
		Status  string
		Message string
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf(`MgmtServer: request for current status`)
		b, err := json.Marshal(global_proxy)
		if err != nil {
			log.Printf(`MgmtServer: error marshalling JSON %v`, err)
			http.Error(w, "Error marshalling to JSON for /", http.StatusInternalServerError)
		}
		fmt.Fprintf(w, string(b))
	})

	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		log.Printf(`MgmtServer: request to refresh`)
		myIpAddress := GetOutboundIP()
		log.Printf("MgmtServer: Refresh, My IP address is %s", myIpAddress)
		global_proxy.UpdateIp(myIpAddress.String())
		log.Printf("MgmtServer: Refresh, updated IP address")

		global_proxy.SearchDomain = GetSearchDomain()

		detected := true
		pac, err := GetWpad("wpad", global_proxy.SearchDomain)
		if err != nil {
			log.Printf(`MgmtServer: error getting wpad = %v`, err)
			log.Printf(`MgmtServer: all connections will be direct`)
			detected = false
		}
		global_proxy.UpdatePac(pac, detected)
		log.Printf("MgmtServer: Refresh, updated PAC details")

		res := &resp{"ok", "refreshed"}
		b, err := json.Marshal(res)
		if err != nil {
			log.Printf(`MgmtServer: error marshalling JSON %v`, err)
			http.Error(w, "Error marshalling to JSON for /refresh", http.StatusInternalServerError)
		}
		fmt.Fprintf(w, string(b))
	})

	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:    fmt.Sprintf(`127.0.0.1:%v`, port),
		Handler: mux,
	}

	return &server
}

func main() {
	// parameters
	proxyPort := flag.Int("proxy", 8080, "Port on which to run the proxy server")
	mgmtPort := flag.Int("mgmt", 9001, "Port on which to run the management server")

	// print the hello messages
	// second parameter is the app version number
	fmt.Printf(motd, 0.1)

	// parse parameters
	flag.Parse()

	// Get my IP address
	myIpAddress := GetOutboundIP()
	log.Printf("Proxy: My IP address is %s", myIpAddress)

	// Get my search domain
	mySearchDomain := GetSearchDomain()
	log.Printf("Proxy: My search domain is: %s", mySearchDomain)

	// get PAC content
	// make a call to http://wpad
	detected := true
	pac, err := GetWpad("wpad", mySearchDomain)
	if err != nil {
		log.Printf(`Proxy: error getting wpad = %v`, err)
		log.Printf(`Proxy: all connections will be direct`)
		detected = false
	}

	// init proxy
	global_proxy = NewProxy(pac, myIpAddress.String(), mySearchDomain, detected)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// mgmt server
	go func() {
		log.Printf(`Proxy: spawn mgmt server, port = %d`, *mgmtPort)
		server := CreateMgmtServer(*mgmtPort)
		server.ListenAndServe()
		wg.Done()
	}()

	// proxy
	go func() {
		log.Printf(`Proxy: spawn proxy server, port = %d`, *proxyPort)
		addr := fmt.Sprintf("127.0.0.1:%d", *proxyPort)
		if err := http.ListenAndServe(addr, global_proxy); err != nil {
			log.Fatalln("ListenAndServe", err)
		}
	}()

	// wait
	wg.Wait()
}
