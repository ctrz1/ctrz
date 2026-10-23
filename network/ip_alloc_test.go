package network

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// isolateState points the allocator at a throwaway state dir.
func isolateState(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("CtrzStateDir() is /var/lib/ctrz for root; refusing to touch real state. Run this test unprivileged.")
	}
	t.Setenv("HOME", t.TempDir())
}

func lastOctet(t *testing.T, ip string) int {
	t.Helper()
	i := strings.LastIndexByte(ip, '.')
	if i < 0 {
		t.Fatalf("not an IPv4 address: %q", ip)
	}
	n, err := strconv.Atoi(ip[i+1:])
	if err != nil {
		t.Fatalf("not an IPv4 address: %q", ip)
	}
	return n
}

type allocResult struct {
	ip  string
	err error
}

// assignOrTimeout calls AssignContIP but gives up waiting after d. A hung
// call leaves its goroutine spinning until the test binary exits.
func assignOrTimeout(d time.Duration) (allocResult, bool) {
	ch := make(chan allocResult, 1)
	go func() {
		ip, err := AssignContIP()
		ch <- allocResult{ip, err}
	}()
	select {
	case r := <-ch:
		return r, true
	case <-time.After(d):
		return allocResult{}, false
	}
}

// 10.200.1.0 is the network, .1 the bridge/gateway, .255 the broadcast address.
func TestRandIPOnlyReturnsUsableHostAddresses(t *testing.T) {
	for i := 0; i < 20000; i++ {
		ip, _ := randIP(nil)
		if !strings.HasPrefix(ip, "10.200.1.") {
			t.Fatalf("randIP returned %q, outside 10.200.1.0/24", ip)
		}
		if n := lastOctet(t, ip); n < 2 || n > 254 {
			t.Fatalf("randIP returned %q: last octet %d is the network, gateway or broadcast address (usable range is 2-254)", ip, n)
		}
	}
}

func TestAssignContIPNeverReusesAnAddress(t *testing.T) {
	isolateState(t)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		ip, err := AssignContIP()
		if err != nil {
			t.Fatalf("allocation %d: %v", i, err)
		}
		if seen[ip] {
			t.Fatalf("allocation %d returned %s which was already handed out", i, ip)
		}
		seen[ip] = true
	}
}

func TestAssignContIPReportsExhaustionInsteadOfHanging(t *testing.T) {
	isolateState(t)
	seen := map[string]bool{}
	// A /24 has at most 254 host addresses, so within 300 calls the allocator
	// must either have returned an error or have been caught handing out dups.
	for i := 0; i < 300; i++ {
		r, ok := assignOrTimeout(3 * time.Second)
		if !ok {
			t.Fatalf("AssignContIP hung on call #%d after handing out %d distinct addresses: "+
				"randIP loops forever once the pool is full; it should return an error", i+1, len(seen))
		}
		if r.err != nil {
			if len(seen) < 200 {
				t.Fatalf("AssignContIP failed after only %d allocations: %v", len(seen), r.err)
			}
			return // exhaustion reported as an error: this is the desired behaviour
		}
		if seen[r.ip] {
			t.Fatalf("call #%d handed out %s a second time", i+1, r.ip)
		}
		seen[r.ip] = true
	}
	t.Fatalf("handed out %d unique addresses from a /24 and never reported exhaustion", len(seen))
}

// Characterisation test: releasing an address works and removes only that one.
// This passes today; it pins the behaviour so a rewrite of the allocator keeps it.
func TestRemoveContIPReleasesOnlyThatAddress(t *testing.T) {
	isolateState(t)
	a, err := AssignContIP()
	if err != nil {
		t.Fatal(err)
	}
	b, err := AssignContIP()
	if err != nil {
		t.Fatal(err)
	}
	if err := RemoveContIP(a); err != nil {
		t.Fatalf("RemoveContIP(%s): %v", a, err)
	}

	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, ".local", "share", "ctrz", "containers", "IP"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Fields(string(data))
	if len(lines) != 1 || lines[0] != b {
		t.Fatalf("after releasing %s the IP file should contain only %s, got %q", a, b, lines)
	}
}
