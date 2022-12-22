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
	result, _ := RunWpadPac(pac, ip, "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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
	result, _ := RunWpadPac(pac, "", "", "")
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

	result, _ := RunWpadPac(pac, "", "", "")
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

	result, _ := RunWpadPac(pac, "", "", "")
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

	result, _ := RunWpadPac(pac, "", "", "")
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

	result, _ := RunWpadPac(pac, "", "", "")
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

	result, _ := RunWpadPac(pac, "", "", "")
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

	result, _ := RunWpadPac(pac, "", "", "")
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

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeCurrentYear(t *testing.T) {
	pac := `
	var year = new Date().getFullYear();

	function FindProxyForURL(url, host) {
		return dateRange(year) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeCurrentMonth(t *testing.T) {
	pac := `
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[new Date().getMonth()];

	function FindProxyForURL(url, host) {
		return dateRange(month) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeDayAndMonth(t *testing.T) {
	pac := `
	var today = new Date().getDate();
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[new Date().getMonth()];

	function FindProxyForURL(url, host) {
		return dateRange(today, month) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeDayAndMonthGMT(t *testing.T) {
	pac := `
	var today = new Date().getUTCDate();
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[new Date().getUTCMonth()];

	function FindProxyForURL(url, host) {
		return dateRange(today, month, "GMT") ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeMonthRange(t *testing.T) {
	pac := `
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[new Date().getMonth()];
	var nextmonth = months[(new Date().getMonth() + 1) % 12];

	function FindProxyForURL(url, host) {
		return dateRange(month, nextmonth) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeMonthRangeUTC(t *testing.T) {
	pac := `
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[new Date().getUTCMonth()];
	var nextmonth = months[(new Date().getUTCMonth() + 1) % 12];

	function FindProxyForURL(url, host) {
		return dateRange(month, nextmonth, "GMT") ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeMonthRangeNoMatch(t *testing.T) {
	pac := `
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[(new Date().getMonth() + 1) % 12];
	var nextmonth = months[(new Date().getMonth() + 2) % 12];

	function FindProxyForURL(url, host) {
		return dateRange(month, nextmonth) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacDateRangeDayAndMonthRange(t *testing.T) {
	pac := `
	var d = new Date();
	var today = d.getDate();
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[d.getMonth()];

	var endofrange = new Date(d);
	endofrange.setDate(endofrange.getDate() + 7);

	function FindProxyForURL(url, host) {
		return dateRange(today, month, endofrange.getDate(), months[endofrange.getMonth()]) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeDayAndMonthRangeGMT(t *testing.T) {
	pac := `
	var d = new Date();
	var today = d.getUTCDate();
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[d.getUTCMonth()];

	var endofrange = new Date(d);
	endofrange.setDate(endofrange.getDate() + 7);

	function FindProxyForURL(url, host) {
		return dateRange(today, month, endofrange.getUTCDate(), months[endofrange.getUTCMonth()], "GMT") ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacDateRangeDayAndMonthRangeNegative(t *testing.T) {
	pac := `
	var d = new Date();
	d.setDate(d.getDate() + 2);
	var today = d.getDate();
	var months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
	var month = months[d.getMonth()];

	var endofrange = new Date(d);
	endofrange.setDate(endofrange.getDate() + 7);

	function FindProxyForURL(url, host) {
		return dateRange(today, month, endofrange.getDate(), months[endofrange.getMonth()]) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "false" {
		t.Fatalf(`Match came back as true, want false`)
	}
}

func TestPacTimeRangeCurrentHour(t *testing.T) {
	pac := `
	var hour = new Date().getHours();

	function FindProxyForURL(url, host) {
		return timeRange(hour) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacTimeRangeCurrentHourGMT(t *testing.T) {
	pac := `
	var hour = new Date().getUTCHours();

	function FindProxyForURL(url, host) {
		return timeRange(hour, "GMT") ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacTimeRangeCurrentHourRange(t *testing.T) {
	pac := `
	var now = new Date();
	var hour = now.getHours();
	var inOneHour = new Date(now);
	inOneHour.setHours(hour + 1);
	var nexthour = inOneHour.getHours();

	function FindProxyForURL(url, host) {
		return timeRange(hour, nexthour) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacTimeRangeCurrentHourRangeGMT(t *testing.T) {
	pac := `
	var now = new Date();
	var hour = now.getUTCHours();
	var inOneHour = new Date(now);
	inOneHour.setHours(now.getHours() + 1);
	var nexthour = inOneHour.getUTCHours();

	function FindProxyForURL(url, host) {
		console.log("hour = " + hour);
		return timeRange(hour, nexthour, "GMT") ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacTimeRangeCurrentHourMinRange(t *testing.T) {
	pac := `
	var now = new Date();
	var hour = now.getHours();
	var min = now.getMinutes();
	var in15min = new Date(now);
	in15min.setMinutes(min + 15);


	function FindProxyForURL(url, host) {
		return timeRange(hour, min, in15min.getHours(), in15min.getMinutes()) ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}

func TestPacTimeRangeCurrentHourMinRangeGMT(t *testing.T) {
	pac := `
	var now = new Date();
	var hour = now.getUTCHours();
	var min = now.getUTCMinutes();
	var in15min = new Date(now);
	in15min.setMinutes(now.getMinutes() + 15);


	function FindProxyForURL(url, host) {
		return timeRange(hour, min, in15min.getUTCHours(), in15min.getUTCMinutes(), "GMT") ? "true" : "false";
	}
	`

	result, _ := RunWpadPac(pac, "", "", "")
	if result != "true" {
		t.Fatalf(`Match came back as false, want true`)
	}
}
