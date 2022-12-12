package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
	pac string
	ip  string
}

func NewProxy(pac string, ip string) *proxy {
	return &proxy{pac, ip}
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Printf(`ServeHTTP: %v from %v for %v`, req.Method, req.RemoteAddr, req.URL)

	if req.Method == http.MethodConnect {
		log.Printf(`ServeHTTP: this is a tunnel request for port = %v`, req.URL.Port())

		expandedUrl := fmt.Sprintf(`https:%v`, req.URL.String())
		result := RunWpadPac(p.pac, p.ip, expandedUrl, req.Host)
		log.Printf(`ServeHTTP: tunnel, result from pac script = %v`, result)

		dest_conn, err := net.DialTimeout("tcp", req.Host, 10*time.Second)
		if err != nil {
			http.Error(wr, err.Error(), http.StatusServiceUnavailable)
			return
		}
		wr.WriteHeader(http.StatusOK)
		hijacker, ok := wr.(http.Hijacker)
		if !ok {
			http.Error(wr, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		client_conn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(wr, err.Error(), http.StatusServiceUnavailable)
		}
		go transfer(dest_conn, client_conn)
		go transfer(client_conn, dest_conn)
	} else {

		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			http.Error(wr, `Protocol scheme not supported`, http.StatusBadRequest)
			log.Printf(`ServeHTTP: protocal scheme %v is not supported`, req.URL.Scheme)
			return
		}

		client := &http.Client{}

		//http://golang.org/src/pkg/net/http/client.go
		req.RequestURI = ""

		delHopHeaders(req.Header)

		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			appendHostToXForwardHeader(req.Header, clientIP)
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(wr, "Server Error", http.StatusInternalServerError)
			log.Fatal("ServeHTTP:", err)
		}
		defer resp.Body.Close()

		log.Printf(`ServeHTTP: client %v, remote %v, status %v`, req.RemoteAddr, req.URL, resp.Status)

		delHopHeaders(resp.Header)

		copyHeader(wr.Header(), resp.Header)
		wr.WriteHeader(resp.StatusCode)
		io.Copy(wr, resp.Body)
	}
}
