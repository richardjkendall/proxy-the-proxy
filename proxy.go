package main

/*
 * This code is a mixture of this https://gist.github.com/yowu/f7dc34bd4736a65ff28d
 * and this https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c
 */

import (
	"bufio"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// metrics
var (
	totalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "proxy_total_requests",
		Help: "Total requests which have hit the proxy",
	})
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
	Pac      string
	Ip       string
	Detected bool
	cache    *cache
}

func (p *proxy) UpdateIp(ip string) {
	p.Ip = ip
}

func NewProxy(pac string, ip string, detected bool) *proxy {
	c := NewCache()
	return &proxy{pac, ip, detected, c}
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func ConnectUpstream(endpoint string, host string) (net.Conn, error) {
	log.Printf(`ConnectUpstream: connecting to %v for host %v`, endpoint, host)
	conn, err := net.DialTimeout("tcp", endpoint, 10*time.Second)
	if err != nil {
		log.Printf(`ConnectUpstream: error connecting: %v`, err)
		return nil, err
	}
	connectString := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n", host)
	fmt.Fprintf(conn, connectString)
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Printf(`ConnectUpstream: error getting status from upstream: %v`, err)
		return nil, err
	}
	if strings.Contains(status, "200") {
		log.Printf(`ConnectUpstream: got 200 OK from upstream`)
		return conn, nil
	}
	log.Printf(`ConnectUpstream: did not get 200 OK, instead got = %v`, status)
	return nil, errors.New("ConnectUpstream: did not get 200 okay")
}

func GetProxyAddress(result string) (string, error) {
	// split string
	proxies := strings.Split(result, ";")
	if proxies[0] == "DIRECT" {
		return "DIRECT", nil
	} else if strings.HasPrefix(proxies[0], "PROXY") {
		bits := strings.Split(proxies[0], " ")
		if len(bits) > 1 {
			return bits[1], nil
		}
	}
	return "", errors.New(fmt.Sprintf("Could not find valid proxy address in %v", result))
}

func GetUrlHash(url string, ip string) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(ip))
	hasher.Write([]byte(url))
	return hasher.Sum(nil)
}

func (p *proxy) LookupProxy(url url.URL) string {
	result := "DIRECT"
	log.Printf(`LookupProxy: looking up proxy for %v`, url.String())
	urlString := url.String()
	host, port, err := net.SplitHostPort(url.Host)
	if err != nil {
		log.Printf(`LookupProxy: error getting host and port from URL so will go with value from URL, %v`, err)
		host = url.Host
		port = "80"
	}
	if !strings.HasPrefix(urlString, "http") {
		if port == "443" {
			urlString = fmt.Sprintf(`https:%v`, urlString)
		} else {
			urlString = fmt.Sprintf(`http:%v`, urlString)
		}
		log.Printf(`LookupProxy: expanding URL as it was missing the scheme, expanded URL = %v`, urlString)
	}
	urlhash := GetUrlHash(urlString, p.Ip)
	cacheValue, err := p.cache.CheckForVal(urlhash)
	if err == nil {
		log.Printf(`LookupProxy: got value from cache = %v`, cacheValue)
		return string(cacheValue)
	}
	result = RunWpadPac(p.Pac, p.Ip, urlString, host)
	proxyaddress, err := GetProxyAddress(result)
	if err != nil {
		log.Printf(`LookupProxy: see above, error getting proxy address, will go direct`)
		return "DIRECT"
	} else {
		log.Printf(`LookupProxy: returning %v`, proxyaddress)
		p.cache.AddVal(urlhash, []byte(proxyaddress))
		return proxyaddress
	}
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Printf(`ServeHTTP: %v from %v for %v`, req.Method, req.RemoteAddr, req.URL)
	totalRequests.Inc()

	if req.Method == http.MethodConnect {
		log.Printf(`ServeHTTP: this is a tunnel request for port = %v`, req.URL.Port())

		result := "DIRECT"
		if p.Detected {
			log.Printf(`ServeHTTP: tunnel: looking up proxy...`)
			result = p.LookupProxy(*req.URL)
		}

		endpoint := req.Host
		var dest_conn *net.Conn
		if result != "DIRECT" {
			endpoint = result
			log.Printf(`ServeHTTP: tunnel, connection to %v will go via %v`, req.URL, endpoint)

			conn, err := ConnectUpstream(endpoint, req.Host)
			if err != nil {
				http.Error(wr, "Upstream connection failed", http.StatusInternalServerError)
				return
			}
			dest_conn = &conn
		} else {
			log.Printf(`ServeHTTP: going direct for %v`, endpoint)
			conn, err := net.DialTimeout("tcp", endpoint, 10*time.Second)
			if err != nil {
				http.Error(wr, "Upstream connection failed", http.StatusInternalServerError)
				return
			}
			dest_conn = &conn
		}

		// send downstream status OK
		wr.WriteHeader(http.StatusOK)
		// hijack downstream
		hijacker, ok := wr.(http.Hijacker)
		if !ok {
			http.Error(wr, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		client_conn, _, err := hijacker.Hijack()
		if err != nil {
			log.Printf(`ServeHTTP: Error after connection hijack: %v`, err)
			http.Error(wr, err.Error(), http.StatusServiceUnavailable)
		}
		// wire together the connections
		go transfer(*dest_conn, client_conn)
		go transfer(client_conn, *dest_conn)
	} else {

		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			http.Error(wr, `Protocol scheme not supported`, http.StatusBadRequest)
			log.Printf(`ServeHTTP: protocol scheme %v is not supported`, req.URL.Scheme)
			return
		}

		client := &http.Client{
			Transport: &http.Transport{
				Proxy: func(r *http.Request) (*url.URL, error) {
					result := "DIRECT"
					if p.Detected {
						log.Printf(`ServeHTTP: looking up proxy...`)
						result = p.LookupProxy(*r.URL)

					}
					if result == "DIRECT" {
						return nil, nil
					} else {
						proxyUrl := fmt.Sprintf(`http://%v`, result)
						proxy, err := url.Parse(proxyUrl)
						if err != nil {
							log.Printf(`ServeHTTP: got error while parsing proxy URL %v`, err)
							return nil, err
						}
						log.Printf(`ServeHTTP: using proxy %v`, proxy)
						return proxy, nil
					}
				},
			},
		}

		//http://golang.org/src/pkg/net/http/client.go
		req.RequestURI = ""

		delHopHeaders(req.Header)

		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			appendHostToXForwardHeader(req.Header, clientIP)
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(wr, "Server Error", http.StatusInternalServerError)
			//log.Fatal("ServeHTTP:", err)
		}
		defer resp.Body.Close()

		log.Printf(`ServeHTTP: client %v, remote %v, status %v`, req.RemoteAddr, req.URL, resp.Status)

		delHopHeaders(resp.Header)

		copyHeader(wr.Header(), resp.Header)
		wr.WriteHeader(resp.StatusCode)
		io.Copy(wr, resp.Body)
	}
}
