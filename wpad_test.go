package main

import "testing"

/*
 * Requirements from here
 * https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file
 *
 * Also used this implementation for inspiration
 * https://source.chromium.org/chromium/chromium/src/+/main:services/proxy_resolver/pac_js_library.h
 */

func TestPacMyIpAddress(t *testing.T) {
	ip := "10.10.10.10"
	pac := `
	function FindProxyForURL(url, host) {
		return myIpAddress();
	}
	`
	result := RunWpadPac(pac, ip)
	if result != ip {
		t.Fatalf(`IP came back as %v, want %v`, result, ip)
	}
}

func TestPacIsNotPlainHostName(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isPlainHostName("www.mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Check for isPlainHostName match came back as true, want false`)
	}
}

func TestPacIsPlainHostName(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isPlainHostName("www")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Check for isPlainHostName match came back as false, want true`)
	}
}

func TestPacIsPlainHostNameInvalidCall(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isPlainHostName()) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Check for isPlainHostName match came back as true, want false`)
	}
}

func TestPacDnsDomainIs(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(dnsDomainIs("www.mozilla.org", ".mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Domain match came back as false, want true`)
	}
}

func TestPacDnsDomainIsNot(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(dnsDomainIs("www", ".mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Domain match came back as true, want false`)
	}
}

func TestPacDnsDomainIsInvalidCall(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return dnsDomainIs() ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Domain match came back as true, want false`)
	}
}

func TestPacDnsLocalHostOrDomainIsExact(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(localHostOrDomainIs("www.mozilla.org", "www.mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDnsLocalHostOrDomainIsHostMatch(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(localHostOrDomainIs("www", "www.mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDnsLocalHostOrDomainIsDomainMismatch(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(localHostOrDomainIs("www.google.com", "www.mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacDnsLocalHostOrDomainIsHostMismatch(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(localHostOrDomainIs("home.mozilla.org", "www.mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacDnsIsResolvable(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isResolvable("www.mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDnsIsNotResolvable(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isResolvable("blahblahblah.mozilla.org")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacDnsIsResolvableInvalidCall(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isResolvable()) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacDnsIsInNet(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isInNet("10.0.1.2", "10.0.1.1", "255.255.255.0")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDnsIsNotInNet(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isInNet("10.0.2.2", "10.0.1.1", "255.255.255.0")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacDnsIsInNetInvalidCall(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isInNet("10.0.1.2")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacDnsIsInNetInvalidCall2(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		if(isInNet("10.0.1.256", "10.0.1.1", "255.255.255.0")) {
			return "true";
		} else {
			return "false";
		}
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacConvertAddr(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return convert_addr("104.16.41.2") == 1745889538 ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacConvertAddrInvalidIp(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return convert_addr("300.16.41.2") == 0 ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacConvertAddrInvalidCall(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return convert_addr() == 0 ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDnsDomainLevels2(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return dnsDomainLevels("test.test.test") == 2 ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDnsDomainLevels0(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return dnsDomainLevels("test") == 0 ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDnsDomainLevelsInvalidCall(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return dnsDomainLevels() == 0 ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacShExpMatchWildcardTrue(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return shExpMatch("http://home.netscape.com/people/ari/index.html", "*/ari/*") ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacShExpMatchWildcardFalse(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return shExpMatch("http://home.netscape.com/people/montulli/index.html", "*/ari/*") ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacShExpMatchQuestionmarkTrue(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return shExpMatch("http://home.netscape.com/people/r/index.html", "*/?/*") ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacShExpMatchQuestionmarkFalse(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return shExpMatch("http://home.netscape.com/people/rr/index.html", "*/?/*") ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacShExpMatchWithDotTrue(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return shExpMatch("http://home.netscape.com/people/rr/index.html", "*.html") ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacShExpMatchWithDotFalse(t *testing.T) {
	pac := `
	function FindProxyForURL(url, host) {
		return shExpMatch("http://home.netscape.com/people/rr/index.html", "*.php") ? "true" : "false";
	}
	`
	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacWeekdayRangeTodayMatch(t *testing.T) {
	pac := `
	var wdays = ["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"];
	var today = wdays[new Date().getDay()];

	function FindProxyForURL(url, host) {
		return weekdayRange(today) ? "true" : "false";
	}
	`

	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacWeekdayRangeTomorrow(t *testing.T) {
	pac := `
	var wdays = ["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"];
	var tomorrow = wdays[(new Date().getDay() + 1) % 7];

	function FindProxyForURL(url, host) {
		return weekdayRange(tomorrow) ? "true" : "false";
	}
	`

	result := RunWpadPac(pac, "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacWeekdayRangeTodayUTCMatch(t *testing.T) {
	pac := `
	var wdays = ["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"];
	var today = wdays[new Date().getUTCDay()];

	function FindProxyForURL(url, host) {
		return weekdayRange(today, "GMT") ? "true" : "false";
	}
	`

	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacWeekdayRangeTodayTomorrowMatch(t *testing.T) {
	pac := `
	var wdays = ["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"];
	var today = wdays[new Date().getDay()];
	var tomorrow = wdays[(new Date().getDay() + 1) % 7];

	function FindProxyForURL(url, host) {
		return weekdayRange(today, tomorrow) ? "true" : "false";
	}
	`

	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacWeekdayRangeTodayTomorrowUTCMatch(t *testing.T) {
	pac := `
	var wdays = ["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"];
	var today = wdays[new Date().getUTCDay()];
	var tomorrow = wdays[(new Date().getUTCDay() + 1) % 7];

	function FindProxyForURL(url, host) {
		return weekdayRange(today, tomorrow, "GMT") ? "true" : "false";
	}
	`

	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeToday(t *testing.T) {
	pac := `
	var today = new Date().getDate();

	function FindProxyForURL(url, host) {
		return dateRange(today) ? "true" : "false";
	}
	`

	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeTodayUTC(t *testing.T) {
	pac := `
	var today = new Date().getUTCDate();

	function FindProxyForURL(url, host) {
		return dateRange(today, "GMT") ? "true" : "false";
	}
	`

	result := RunWpadPac(pac, "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}
