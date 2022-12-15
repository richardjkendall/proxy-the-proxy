package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

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
		res := &resp{"ok", "refreshed"}
		b, err := json.Marshal(res)
		if err != nil {
			log.Printf(`MgmtServer: error marshalling JSON %v`, err)
			http.Error(w, "Error marshalling to JSON for /refresh", http.StatusInternalServerError)
		}
		fmt.Fprintf(w, string(b))
	})

	server := http.Server{
		Addr:    fmt.Sprintf(`127.0.0.1:%v`, port),
		Handler: mux,
	}

	return &server
}

func main() {

	// Get my IP address
	myIpAddress := GetOutboundIP()
	log.Printf("Proxy: My IP address is %s", myIpAddress)

	// get PAC content
	detected := true
	pac, err := GetWpad("wpad")
	if err != nil {
		log.Printf(`Proxy: error getting wpad = %v`, err)
		log.Printf(`Proxy: all connections will be direct`)
		detected = false
	}

	// init proxy
	global_proxy = NewProxy(pac, myIpAddress.String(), detected)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// mgmt server
	go func() {
		log.Printf(`Proxy: spawn mgmt server`)
		server := CreateMgmtServer(9001)
		server.ListenAndServe()
		wg.Done()
	}()

	// proxy
	go func() {
		log.Printf(`Proxy: spawn proxy server`)
		addr := "127.0.0.1:8080"
		if err := http.ListenAndServe(addr, global_proxy); err != nil {
			log.Fatalln("ListenAndServe", err)
		}
	}()

	// wait
	wg.Wait()
}
