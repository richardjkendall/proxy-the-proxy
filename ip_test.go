package main

import "testing"

func TestCountBitsInMaskSlash32(t *testing.T) {
	count, err := CountBitsInMask("255.255.255.255")
	if err != nil {
		t.Fatalf("Error calling CountBitsInMask: %v", err)
	}
	if count != 32 {
		t.Fatalf("Got mask length of %v, but expected 32", count)
	}
}

func TestCountBitsInMaskSlash24(t *testing.T) {
	count, err := CountBitsInMask("255.255.255.0")
	if err != nil {
		t.Fatalf("Error calling CountBitsInMask: %v", err)
	}
	if count != 24 {
		t.Fatalf("Got mask length of %v, but expected 24", count)
	}
}

func TestCountBitsInMaskSlash16(t *testing.T) {
	count, err := CountBitsInMask("255.255.0.0")
	if err != nil {
		t.Fatalf("Error calling CountBitsInMask: %v", err)
	}
	if count != 16 {
		t.Fatalf("Got mask length of %v, but expected 16", count)
	}
}

func TestCountBitsInMaskSlash21(t *testing.T) {
	count, err := CountBitsInMask("255.255.248.0")
	if err != nil {
		t.Fatalf("Error calling CountBitsInMask: %v", err)
	}
	if count != 21 {
		t.Fatalf("Got mask length of %v, but expected 21", count)
	}
}

func TestGetNetFromIPandMask(t *testing.T) {
	cidr, err := GetNetFromIPandMask("10.0.1.20", "255.255.255.0")
	if err != nil {
		t.Fatalf("Error calling GetNetFromIPandMask: %v", err)
	}
	if cidr.String() != "10.0.1.0/24" {
		t.Fatalf("Got CIDR of %v, but expected 10.0.1.0/24", cidr)
	}
}

func TestIsIpInRange(t *testing.T) {
	inrange, err := IsIpInRange("10.0.1.5", "10.0.1.1", "255.255.255.0")
	if err != nil {
		t.Fatalf("Error calling IsIpInRange: %v", err)
	}
	if !inrange {
		t.Fatalf("Got false when expected true")
	}
}

func TestIsIpNotInRange(t *testing.T) {
	inrange, err := IsIpInRange("10.0.2.5", "10.0.1.1", "255.255.255.0")
	if err != nil {
		t.Fatalf("Error calling IsIpInRange: %v", err)
	}
	if inrange {
		t.Fatalf("Got true when expected false")
	}
}
