package ipallocator

import (
	"testing"
	"net"
)

func TestAllocate(t *testing.T) {
	testCases := []struct {
		name             string
		cidr             string
		free             int
		released         string
		outOfRange1      string
		outOfRange2      string
		outOfRange3      string
		alreadyAllocated string
	}{
		{
			name:             "IPv4",
			cidr:             "192.168.1.0/24",
			free:             254,
			released:         "192.168.1.5",
			outOfRange1:      "192.168.0.1",
			outOfRange2:      "192.168.1.0",
			outOfRange3:      "192.168.1.255",
			alreadyAllocated: "192.168.1.1",
		},
		{
			name:             "IPv6",
			cidr:             "2001:db8:1::/48",
			free:             65534,
			released:         "2001:db8:1::5",
			outOfRange1:      "2001:db8::1",
			outOfRange2:      "2001:db8:1::",
			outOfRange3:      "2001:db8:1::ffff",
			alreadyAllocated: "2001:db8:1::1",
		},
	}
	for _, tc := range testCases {
		_, cidr, err := net.ParseCIDR(tc.cidr)
		if err != nil {
			t.Fatal(err)
		}
		r := NewCIDRRange(cidr)
		t.Logf("base: %v", r.base.Bytes())
		if f := r.Free(); f != tc.free {
			t.Errorf("Test %s unexpected free %d", tc.name, f)
		}
		if f := r.Used(); f != 0 {
			t.Errorf("Test %s unexpected used %d", tc.name, f)
		}
		count := 0
		for r.Free() > 0 {
			ip, err := r.AllocateNext()
			if err != nil {
				t.Fatalf("Test %s error @ %d: %v", tc.name, count, err)
			}
			count++
			if !cidr.Contains(ip) {
				t.Fatalf("Test %s allocated %s which is outside of %s", tc.name, ip, cidr)
			}
		}
		if _, err := r.AllocateNext(); err != ErrFull {
			t.Fatal(err)
		}

		released := net.ParseIP(tc.released)
		if err := r.Release(released); err != nil {
			t.Fatal(err)
		}
		if f := r.Free(); f != 1 {
			t.Errorf("Test %s unexpected free %d", tc.name, f)
		}
		if f := r.Used(); f != (tc.free - 1) {
			t.Errorf("Test %s unexpected free %d", tc.name, f)
		}
		ip, err := r.AllocateNext()
		if err != nil {
			t.Fatal(err)
		}
		if !released.Equal(ip) {
			t.Errorf("Test %s unexpected %s : %s", tc.name, ip, released)
		}

		if err := r.Release(released); err != nil {
			t.Fatal(err)
		}
		err = r.Allocate(net.ParseIP(tc.outOfRange1))
		if _, ok := err.(*ErrNotInRange); !ok {
			t.Fatal(err)
		}
		if err := r.Allocate(net.ParseIP(tc.alreadyAllocated)); err != ErrAllocated {
			t.Fatal(err)
		}
		err = r.Allocate(net.ParseIP(tc.outOfRange2))
		if _, ok := err.(*ErrNotInRange); !ok {
			t.Fatal(err)
		}
		err = r.Allocate(net.ParseIP(tc.outOfRange3))
		if _, ok := err.(*ErrNotInRange); !ok {
			t.Fatal(err)
		}
		if f := r.Free(); f != 1 {
			t.Errorf("Test %s unexpected free %d", tc.name, f)
		}
		if f := r.Used(); f != (tc.free - 1) {
			t.Errorf("Test %s unexpected free %d", tc.name, f)
		}
		if err := r.Allocate(released); err != nil {
			t.Fatal(err)
		}
		if f := r.Free(); f != 0 {
			t.Errorf("Test %s unexpected free %d", tc.name, f)
		}
		if f := r.Used(); f != tc.free {
			t.Errorf("Test %s unexpected free %d", tc.name, f)
		}
	}
}
