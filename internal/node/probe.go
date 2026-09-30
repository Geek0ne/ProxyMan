package node

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

// DelayResult holds a measured round-trip time for one node
type DelayResult struct {
	Name    string
	Server  string
	Port    int
	DelayMS int
	OK      bool
	Err     string
}

// Default probe settings, overridable via the command line.
const (
	// DefaultProbeTimeout bounds a single node's reachability check.
	DefaultProbeTimeout = 5 * time.Second
	// DefaultProbeJobs is how many nodes are probed at once.
	DefaultProbeJobs = 8
	// MaxProbeJobs caps concurrency so a large node list cannot exhaust
	// file descriptors on the host.
	MaxProbeJobs = 64
)

// Prober measures node reachability with bounded concurrency
type Prober struct {
	Timeout time.Duration
	Jobs    int
}

// NewProber returns a prober with the package defaults
func NewProber() *Prober {
	return &Prober{Timeout: DefaultProbeTimeout, Jobs: DefaultProbeJobs}
}

// NewProberWith returns a prober using caller-supplied settings, falling back
// to the defaults for non-positive values and clamping Jobs to MaxProbeJobs.
func NewProberWith(timeout time.Duration, jobs int) *Prober {
	if timeout <= 0 {
		timeout = DefaultProbeTimeout
	}
	switch {
	case jobs < 1:
		jobs = DefaultProbeJobs
	case jobs > MaxProbeJobs:
		jobs = MaxProbeJobs
	}
	return &Prober{Timeout: timeout, Jobs: jobs}
}

// Probe measures all supplied nodes and returns results sorted by latency
func (p *Prober) Probe(nodes []*Node) []DelayResult {
	if p.Jobs < 1 {
		p.Jobs = 8
	}
	if p.Timeout <= 0 {
		p.Timeout = 5 * time.Second
	}

	results := make([]DelayResult, len(nodes))
	sem := make(chan struct{}, p.Jobs)
	var wg sync.WaitGroup

	for i, n := range nodes {
		wg.Add(1)
		go func(i int, n *Node) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = p.probeOne(n)
		}(i, n)
	}
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].OK != results[j].OK {
			return results[i].OK
		}
		return results[i].DelayMS < results[j].DelayMS
	})
	return results
}

func (p *Prober) probeOne(n *Node) DelayResult {
	res := DelayResult{Name: n.Name, Server: n.Server, Port: n.Port}
	if n.Server == "" || n.Port <= 0 {
		res.Err = "incomplete address"
		return res
	}

	addr := net.JoinHostPort(n.Server, strconv.Itoa(n.Port))
	dialer := &net.Dialer{Timeout: p.Timeout}

	// TCP handshake is the cheapest meaningful reachability signal and works
	// for every protocol ProxyMan supports.
	start := time.Now()
	conn, err := dialer.DialContext(context.Background(), "tcp", addr)
	elapsed := time.Since(start)

	if err != nil {
		res.Err = trimError(err.Error())
		return res
	}
	_ = conn.Close()

	res.OK = true
	res.DelayMS = int(elapsed.Milliseconds())
	return res
}

func trimError(msg string) string {
	const max = 48
	if len(msg) <= max {
		return msg
	}
	return msg[:max] + "…"
}

// Report renders probe results as a fixed-width table
func Report(results []DelayResult) string {
	out := ""
	out += fmt.Sprintf("  %-28s %-24s %8s\n", "NODE", "ENDPOINT", "DELAY")
	out += "  " + repeat('-', 62) + "\n"
	for _, r := range results {
		endpoint := fmt.Sprintf("%s:%d", r.Server, r.Port)
		if len(endpoint) > 24 {
			endpoint = endpoint[:23] + "…"
		}
		if r.OK {
			out += fmt.Sprintf("  %-28s %-24s %7dms\n", trim(r.Name, 28), endpoint, r.DelayMS)
		} else {
			out += fmt.Sprintf("  %-28s %-24s %8s\n", trim(r.Name, 28), endpoint, "unreachable")
		}
	}
	return out
}

func trim(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func repeat(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}
