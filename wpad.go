package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robertkrimen/otto"
)

var (
	scriptExecutions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "proxy_wpad_script_executions",
		Help: "Total times a WPAD script has been executed",
	})
	wpadExecTimeHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "proxy_wpad_exec_seconds",
		Help:    "Histogram of WPAD script execution times in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2},
	})
)

func GetWpadFqdn(host string, searchdomain []string) string {
	// need to go through search domains
	if len(searchdomain) == 0 {
		log.Printf(`GetWpadFqdn: searchdomain length = 0, returning wpad`)
		return host
	} else {
		for _, domain := range searchdomain {
			// try to find wpad
			domainbits := strings.Split(domain, ".")
			for i := 0; i < len(domainbits); i++ {
				fqdn := fmt.Sprintf("%s.%s", host, strings.Join(domainbits[i:], "."))
				log.Printf(`GetWpadFqdn: domain=%s, trying=%s`, domain, fqdn)
				if PerformDNSLookup(fqdn) != false {
					return fqdn
				}
			}
		}
	}
	return host
}

func GetWpad(host string, searchdomain []string) (string, error) {
	// get Wpad domain
	wpad_domain := GetWpadFqdn(host, searchdomain)
	log.Printf(`GetWpad: using domain of %s`, wpad_domain)
	// get PAC file
	resp, err := http.Get(fmt.Sprintf("http://%s/wpad.dat", wpad_domain))
	if err != nil {
		log.Printf(`GetWpad: error on connection: %v`, err)
		return "", err
	}

	// read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf(`GetWpad: error getting response: %v`, err)
		return "", err
	}

	// check response code
	if resp.StatusCode > 299 {
		log.Printf(`GetWpad: got a non 2xx response code: %v`, resp.StatusCode)
		return "", errors.New("GetWpad: got a non 2xx response code")
	}

	sb := string(body)
	return sb, nil
}

func RunWpadPac(pac string, ipaddress string, url string, host string) (string, bool) {
	scriptExecutions.Inc()
	cacheable := true
	vm := otto.New()

	log.Printf(`RunWpadPac: ip = %v, url = %v, host = %v`, ipaddress, url, host)

	// set variables
	vm.Set("INJ_REQ_URL", url)
	vm.Set("INJ_REQ_HOST", host)

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
		log.Printf("shExpMatch: str = %v, shexp = %v", str, shexp)
		pattern := strings.ReplaceAll(shexp, ".", "\\.")
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		pattern = strings.ReplaceAll(pattern, "?", ".")
		log.Printf("shExpMatch: pattern = ^%v$", pattern)
		r, err := regexp.Compile("^" + pattern + "$")
		if err != nil {
			log.Printf("shExpMatch: error compiling re = %v", err)
		}
		match := r.MatchString(str)
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

	vm.Set("timeRange", func(call otto.FunctionCall) otto.Value {
		argc := len(call.ArgumentList)
		log.Printf("timeRange: argc = %v", argc)
		gmt := false
		if argc < 1 {
			// no arguments
			result, _ := vm.ToValue(false)
			return result
		}
		cacheable = false
		// handle GMT
		last, _ := call.Argument(argc - 1).ToString()
		if last == "GMT" {
			log.Printf("dateRange: should be using GMT/UTC")
			gmt = true
			argc--
		}
		currentHour := time.Now().Hour()
		date1 := time.Now()
		date2 := time.Now()
		now := time.Now()
		if gmt {
			currentHour = time.Now().UTC().Hour()
			date1 = date1.In(time.UTC)
			date2 = date1.In(time.UTC)
		}
		if argc == 1 {
			// single argument, this is an hour
			arg0, _ := call.Argument(0).ToInteger()
			log.Printf("dateRange: single argument case, arg0 = %v, currenthour = %v", arg0, currentHour)
			result, _ := vm.ToValue(arg0 == int64(currentHour))
			return result
		}
		if argc == 2 {
			// two arguments, assuming also to be hours
			arg0, _ := call.Argument(0).ToInteger()
			arg1, _ := call.Argument(1).ToInteger()
			log.Printf("dateRange: double argument case, arg0 = %v, arg1 = %v, currenthour = %v", arg0, arg1, currentHour)
			result, _ := vm.ToValue(arg0 <= int64(currentHour) && int64(currentHour) <= arg1)
			return result
		}
		switch argc {
		case 6:
			// six arguments so assumed to be hh mm ss, hh mm ss
			arg2, _ := call.Argument(2).ToInteger()
			arg5, _ := call.Argument(5).ToInteger()
			date1 = UpdateSecondsOfTime(date1, int(arg2))
			date2 = UpdateSecondsOfTime(date2, int(arg5))
			fallthrough
		case 4:
			// four arguments so assumed to be hh mm, hh mm
			middle := argc >> 1
			arg0, _ := call.Argument(0).ToInteger()
			arg1, _ := call.Argument(1).ToInteger()
			date1 = UpdateHoursOfTime(date1, int(arg0))
			date1 = UpdateMinutesOfTime(date1, int(arg1))
			argm, _ := call.Argument(middle).ToInteger()
			argm1, _ := call.Argument(middle + 1).ToInteger()
			date2 = UpdateHoursOfTime(date2, int(argm))
			date2 = UpdateMinutesOfTime(date2, int(argm1))
			log.Printf("dateRange: four argument case, arg0 = %v, arg1 = %v, arg2 = %v, arg3 = %v", arg0, arg1, argm, argm1)
			if middle == 2 {
				date2 = UpdateSecondsOfTime(date2, 59)
			}
			break
		default:
			err := vm.MakeCustomError("timeRange", "bad number of arguments")
			return err
		}
		if gmt {
			now = now.In(time.UTC)
		}
		log.Printf("timeRange: date1 = %v, date2 = %v, GMT = %v, now = %v", date1, date2, gmt, now)
		output := false
		if DateLTE(date1, date2) {
			log.Printf("timeRange: date1 <= date2, date1 <= now = %v, now <= date2 = %v", DateLTE(date1, now), DateLTE(now, date2))
			output = DateLTE(date1, now) && DateLTE(now, date2)
		} else {
			output = DateGTE(date2, now) || DateGTE(now, date1)
		}
		result, _ := vm.ToValue(output)
		return result
	})

	vm.Set("dateRange", func(call otto.FunctionCall) otto.Value {
		months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
		argc := len(call.ArgumentList)
		log.Printf("dateRange: argc = %v", argc)
		gmt := false
		if argc < 1 {
			// no arguments
			result, _ := vm.ToValue(false)
			return result
		}
		cacheable = false
		last, _ := call.Argument(argc - 1).ToString()
		if last == "GMT" {
			log.Printf("dateRange: should be using GMT/UTC")
			gmt = true
			argc--
		}
		// case with a single argument (after checking if we have GMT)
		if argc == 1 {
			arg0, _ := call.Argument(0).ToString()
			log.Printf("dateRange: single argument case, arg0 = %v", arg0)
			a0, err := strconv.Atoi(arg0)
			if err != nil {
				// arg0 is not a number
				// so assume it is a month string

				a0 = indexOf(arg0, months)
				currentmonth := int(time.Now().Month()) - 1
				if gmt {
					currentmonth = int(time.Now().UTC().Month()) - 1
				}
				log.Printf("dateRange: assumed to be a month, monthindex = %v, currentmonth = %v", a0, currentmonth)
				output := a0 == currentmonth
				result, _ := vm.ToValue(output)
				return result
			} else {
				// arg0 is a number
				// if it is less than 32 we assume it is a day
				// otherwise it is a a year
				if a0 < 32 {
					currentday := int(time.Now().Day())
					if gmt {
						currentday = int(time.Now().UTC().Day())
					}
					log.Printf("dateRange: assumed to be a day of the month, day = %v, currentday = %v", a0, currentday)
					output := a0 == currentday
					result, _ := vm.ToValue(output)
					return result
				} else {
					currentyear := int(time.Now().Year())
					if gmt {
						currentyear = int(time.Now().UTC().Year())
					}
					log.Printf("dateRange: assumed to be a year, year = %v, currentyear = %v", a0, currentyear)
					output := a0 == currentyear
					result, _ := vm.ToValue(output)
					return result
				}
			}
		}

		// general case
		year := int(time.Now().Year())
		now := time.Now()
		date1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
		date2 := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.Local)
		adjustmonth := false
		// look at first group of args
		// this is for date 1
		for i := 0; i < (argc >> 1); i++ {
			arg, _ := call.Argument(i).ToString()
			log.Printf("dateRange: loop 1, arg = %v, value = %v", i, arg)
			a, err := strconv.Atoi(arg)
			if err != nil {
				// this is not a number, so assume it is a date string
				a = indexOf(arg, months) + 1
				date1 = UpdateMonthOfDate(date1, a)
			} else {
				if a < 32 {
					// this is a day
					adjustmonth = argc <= 2
					date1 = UpdateDayOfDate(date1, a)
				} else {
					date1 = UpdateYearOfDate(date1, a)
				}
			}
		}
		// now the second group
		// this is for date 2
		for i := (argc >> 1); i < argc; i++ {
			arg, _ := call.Argument(i).ToString()
			log.Printf("dateRange: loop 2, arg = %v, value = %v", i, arg)
			a, err := strconv.Atoi(arg)
			if err != nil {
				// this is not a number, so assume it is a date string
				a = indexOf(arg, months) + 1
				date2 = UpdateMonthOfDate(date2, a)
			} else {
				if a < 32 {
					// this is a day
					adjustmonth = argc <= 2
					date2 = UpdateDayOfDate(date2, a)
				} else {
					date2 = UpdateYearOfDate(date2, a)
				}
			}
		}
		// adjust month to current month if needed
		if adjustmonth {
			date1 = UpdateMonthOfDate(date1, int(now.Month()))
			date2 = UpdateMonthOfDate(date2, int(now.Month()))
		}
		// adjust current time to UTC if needed
		if gmt {
			now = now.In(time.UTC)
		}
		log.Printf("dateRange: date1 = %v, date2 = %v, GMT = %v, now = %v", date1, date2, gmt, now)
		output := false
		if DateLTE(date1, date2) {
			output = DateLTE(date1, now) && DateLTE(now, date2)
		} else {
			output = DateGTE(date2, now) || DateGTE(now, date1)
		}
		result, _ := vm.ToValue(output)
		return result
	})

	vm.Set("weekdayRange", func(call otto.FunctionCall) otto.Value {
		wdays := []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
		if len(call.ArgumentList) == 0 {
			// no arguments
			result, _ := vm.ToValue(false)
			return result
		}
		cacheable = false
		// check if the last argument is 'GMT'
		wday := 0
		gmt := false
		last, _ := call.Argument(len(call.ArgumentList) - 1).ToString()
		if last == "GMT" {
			log.Printf("weekdayRange: GMT is true")
			wday = int(time.Now().UTC().Weekday())
			gmt = true
		} else {
			wday = int(time.Now().Weekday())
		}
		log.Printf("weekdayRange: today index = %v", wday)
		wd1arg, _ := call.Argument(0).ToString()
		wd2arg, _ := call.Argument(1).ToString()
		wd1 := indexOf(wd1arg, wdays)
		log.Printf("weekdayRange: first weekday = %v, index = %v", wd1arg, wd1)
		wd2 := wd1
		if len(call.ArgumentList) == 3 || (len(call.ArgumentList) == 2 && !gmt) {
			wd2 = indexOf(wd2arg, wdays)
			log.Printf("weekdayRange: got a second weekday = %v, index = %v", wd2arg, wd2)
		}
		output := false
		if wd1 == -1 || wd2 == -1 {
			output = false
		} else {
			if wd1 <= wd2 {
				output = (wd1 <= wday && wday <= wd2)
			} else {
				output = (wd2 >= wday || wday >= wd1)
			}
		}
		result, _ := vm.ToValue(output)
		return result

	})

	start := time.Now()
	_, err := vm.Run(fmt.Sprintf(`
	%s

	var output = FindProxyForURL(INJ_REQ_URL, INJ_REQ_HOST);

	`, pac))
	duration := time.Since(start)
	wpadExecTimeHistogram.Observe(duration.Seconds())

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

	return output_string, cacheable
}
