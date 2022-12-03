package main

import "testing"

/*
 * Requirements from here
 * https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file
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
