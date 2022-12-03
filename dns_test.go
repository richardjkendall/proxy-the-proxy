package main

import "testing"

func TestPerformDNSLookup(t *testing.T) {
	lookup := PerformDNSLookup("test-lookup.richardjameskendall.com")
	switch lookup.(type) {
	case string:
		if lookup != "8.8.8.8" {
			t.Fatalf("Got unexpected IP address back = %v", lookup)
		}
	case bool:
		t.Fatalf("Got unexpected false result")
	default:
		t.Fatalf("Got unexpected type of result")
	}
}

func TestPerformDNSLookupFail(t *testing.T) {
	lookup := PerformDNSLookup("test-lookup-fail.richardjameskendall.com")
	switch lookup.(type) {
	case string:
		t.Fatalf("Got unexpected string result = %v", lookup)
	case bool:
		if lookup == true {
			t.Fatalf("Got unexpected true result")
		}
	default:
		t.Fatalf("Got unexpected type of result")
	}
}
