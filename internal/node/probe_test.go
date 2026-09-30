package node

import (
	"net"
	"testing"
	"time"
)

func TestNewProberUsesDefaults(t *testing.T) {
	p := NewProber()
	if p.Timeout != DefaultProbeTimeout {
		t.Errorf("expected default timeout %s, got %s", DefaultProbeTimeout, p.Timeout)
	}
	if p.Jobs != DefaultProbeJobs {
		t.Errorf("expected default jobs %d, got %d", DefaultProbeJobs, p.Jobs)
	}
}

func TestNewProberWithAcceptsOverrides(t *testing.T) {
	p := NewProberWith(2*time.Second, 32)
	if p.Timeout != 2*time.Second {
		t.Errorf("expected timeout 2s, got %s", p.Timeout)
	}
	if p.Jobs != 32 {
		t.Errorf("expected jobs 32, got %d", p.Jobs)
	}
}

func TestNewProberWithFallsBackOnNonPositive(t *testing.T) {
	// Each case passes a valid value for one parameter and an invalid one for
	// the other; only the invalid parameter should fall back to its default.
	cases := []struct {
		name        string
		timeout     time.Duration
		jobs        int
		wantTimeout time.Duration
		wantJobs    int
	}{
		{"zero timeout", 0, 4, DefaultProbeTimeout, 4},
		{"negative timeout", -time.Second, 4, DefaultProbeTimeout, 4},
		{"zero jobs", 2 * time.Second, 0, 2 * time.Second, DefaultProbeJobs},
		{"negative jobs", 2 * time.Second, -3, 2 * time.Second, DefaultProbeJobs},
		{"both invalid", 0, 0, DefaultProbeTimeout, DefaultProbeJobs},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProberWith(tc.timeout, tc.jobs)
			if p.Timeout != tc.wantTimeout {
				t.Errorf("expected timeout %s, got %s", tc.wantTimeout, p.Timeout)
			}
			if p.Jobs != tc.wantJobs {
				t.Errorf("expected jobs %d, got %d", tc.wantJobs, p.Jobs)
			}
		})
	}
}

func TestNewProberWithClampsJobs(t *testing.T) {
	p := NewProberWith(time.Second, MaxProbeJobs+100)
	if p.Jobs != MaxProbeJobs {
		t.Errorf("expected jobs clamped to %d, got %d", MaxProbeJobs, p.Jobs)
	}
}

func TestProbeReachableListener(t *testing.T) {
	// Bind a throwaway listener so the probe has something real to reach.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot bind local listener: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port := 0
	for _, c := range portStr {
		port = port*10 + int(c-'0')
	}

	p := NewProberWith(time.Second, 2)
	results := p.Probe([]*Node{{Name: "local", Server: host, Port: port}})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Fatalf("expected local listener to be reachable, got err=%q", results[0].Err)
	}
}

func TestProbeSortsReachableFirst(t *testing.T) {
	// 127.0.0.1:1 is closed, so the result must be marked unreachable
	// while the ordering contract still holds.
	p := NewProberWith(300*time.Millisecond, 2)
	results := p.Probe([]*Node{
		{Name: "closed", Server: "127.0.0.1", Port: 1},
		{Name: "invalid", Server: "no-such-host.invalid", Port: 443},
	})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i := 1; i < len(results); i++ {
		if results[i-1].OK && !results[i].OK {
			t.Fatalf("reachable results must sort before unreachable ones: %+v", results)
		}
	}
}
