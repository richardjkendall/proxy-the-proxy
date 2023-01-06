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
	"regexp"
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

	totalBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "proxy_total_bytes_served",
		Help: "Total bytes passed through the proxy",
	})

	proxyUpstreamTunnelConnect = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "proxy_upstream_tunnel_connect_seconds",
		Help:    "Histogram of upstream tunnel connect in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2},
	}, []string{"status_code"})

	proxyUpstreamHttp = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "proxy_upstream_http_seconds",
		Help:    "Histogram of upstream HTTP in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2},
	}, []string{"status_code"})

	proxyServeTimeHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "proxy_serve_time_seconds",
		Help:    "Histogram of the time taken to serve proxy requests",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2},
	}, []string{"proxy"})
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
	Pac          string
	Ip           string
	SearchDomain []string
	Detected     bool
	cache        *cache
}

func (p *proxy) UpdateIp(ip string) {
	p.Ip = ip
}

func (p *proxy) UpdatePac(pac string, detected bool) {
	p.Pac = pac
	p.Detected = detected
}

func NewProxy(pac string, ip string, searchdomain []string, detected bool) *proxy {
	c := NewCache()
	return &proxy{pac, ip, searchdomain, detected, c}
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	written, _ := io.Copy(destination, source)
	totalBytes.Add(float64(written))
}

func GetUpstreamStatus(status string) (bool, string, string) {
	e := `HTTP\/1.\d (\d+) (.*)`
	r := regexp.MustCompile(e)
	result := r.FindAllStringSubmatch(status, -1)
	if result == nil {
		return false, "", ""
	} else {
		return true, result[0][1], result[0][2]
	}
}

func ConnectUpstream(endpoint string, host string) (net.Conn, error) {
	start := time.Now()
	log.Printf(`ConnectUpstream: connecting to %v for host %v`, endpoint, host)
	/*
		Trying to start on instrumenting DNS lookups
		dialer := &net.Dialer{
			Resolver: &net.Resolver{

			},
		}*/
	conn, err := net.DialTimeout("tcp", endpoint, 10*time.Second)
	if err != nil {
		log.Printf(`ConnectUpstream: error connecting: %v`, err)
		return nil, err
	}
	connectString := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n", host)
	fmt.Fprintf(conn, connectString)
	status, err := bufio.NewReader(conn).ReadString('\n')
	okay, code, message := GetUpstreamStatus(status)
	if !okay {
		log.Printf(`ConnectUpstream: did not understand upstream response which was = %v`, status)
		return nil, errors.New("ConnectUpstream: did not get 2xx okay")
	} else {
		duration := time.Since(start)
		proxyUpstreamTunnelConnect.WithLabelValues(code).Observe(duration.Seconds())
		if err != nil {
			log.Printf(`ConnectUpstream: error getting status from upstream: %v`, err)
			return nil, err
		}
		if strings.HasPrefix(code, "2") {
			log.Printf(`ConnectUpstream: got 2xx OK from upstream`)
			return conn, nil
		}
		log.Printf(`ConnectUpstream: did not get 2xx OK, instead got = %v, %v`, code, message)
		return nil, errors.New("ConnectUpstream: did not get 200 okay")
	}
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
	result, cacheable := RunWpadPac(p.Pac, p.Ip, urlString, host)
	proxyaddress, err := GetProxyAddress(result)
	if err != nil {
		log.Printf(`LookupProxy: see above, error getting proxy address, will go direct`)
		return "DIRECT"
	} else {
		log.Printf(`LookupProxy: returning %v, cacheable = %v`, proxyaddress, cacheable)
		if cacheable {
			p.cache.AddVal(urlhash, []byte(proxyaddress))
		}
		return proxyaddress
	}
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	start := time.Now()
	log.Printf(`ServeHTTP: %v from %v for %v`, req.Method, req.RemoteAddr, req.URL)
	totalRequests.Inc()

	target := "DIRECT"

	if req.Method == http.MethodConnect {
		log.Printf(`ServeHTTP: this is a tunnel request for port = %v`, req.URL.Port())

		result := "DIRECT"
		if p.Detected {
			log.Printf(`ServeHTTP: tunnel: looking up proxy...`)
			result = p.LookupProxy(*req.URL)
		}
		target = result

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
					target = result
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
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		//http://golang.org/src/pkg/net/http/client.go
		req.RequestURI = ""

		delHopHeaders(req.Header)

		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			appendHostToXForwardHeader(req.Header, clientIP)
		}

		http_start := time.Now()
		resp, err := client.Do(req)
		http_duration := time.Since(http_start)
		proxyUpstreamHttp.WithLabelValues(fmt.Sprint(resp.StatusCode)).Observe(http_duration.Seconds())
		if err != nil {
			http.Error(wr, "Server Error", http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		log.Printf(`ServeHTTP: client %v, remote %v, status %v`, req.RemoteAddr, req.URL, resp.Status)

		delHopHeaders(resp.Header)

		copyHeader(wr.Header(), resp.Header)
		wr.WriteHeader(resp.StatusCode)
		written, _ := io.Copy(wr, resp.Body)
		totalBytes.Add(float64(written))

	}
	duration := time.Since(start)
	proxyServeTimeHistogram.WithLabelValues(target).Observe(duration.Seconds())
}
