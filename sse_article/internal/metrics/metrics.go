package metrics

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Tracker struct {
	mu              sync.Mutex
	latencies       []time.Duration
	totalRequests   uint64
	totalMessages   uint64
	uselessRequests uint64
	activeConns     int64
	maxActiveConns  int64
	totalInc        int64
	incCount        int
}

type Snapshot struct {
	P50             time.Duration
	P95             time.Duration
	P99             time.Duration
	TotalRequests   uint64
	TotalMessages   uint64
	UselessRequests uint64
	ActiveConns     int64
	MaxActiveConns  int64
	MemMB           float64
	TotalInc        int64
}

func New() *Tracker {
	return &Tracker{}
}

func (t *Tracker) RecordLatency(d time.Duration) {
	if d < 0 {
		return
	}
	t.mu.Lock()
	t.latencies = append(t.latencies, d)
	t.mu.Unlock()
}

func (t *Tracker) AddRequest() {
	atomic.AddUint64(&t.totalRequests, 1)
}

func (t *Tracker) AddMessage() {
	atomic.AddUint64(&t.totalMessages, 1)
}

func (t *Tracker) AddUselessRequest() {
	atomic.AddUint64(&t.uselessRequests, 1)
}

func (t *Tracker) IncActive() {
	t.incCount++
	atomic.AddInt64(&t.totalInc, 1)
	atomic.AddInt64(&t.activeConns, 1)
	current := atomic.LoadInt64(&t.activeConns)
	max := atomic.LoadInt64(&t.maxActiveConns)
	if current > max {
		atomic.StoreInt64(&t.maxActiveConns, current)
	}
}

func (t *Tracker) DecActive() {
	atomic.AddInt64(&t.activeConns, -1)
}

func (t *Tracker) Snapshot() Snapshot {
	var latencies []time.Duration
	t.mu.Lock()
	if len(t.latencies) > 0 {
		latencies = make([]time.Duration, len(t.latencies))
		copy(latencies, t.latencies)
	}
	t.mu.Unlock()

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	return Snapshot{
		P50:             percentile(latencies, 0.50),
		P95:             percentile(latencies, 0.95),
		P99:             percentile(latencies, 0.99),
		TotalRequests:   atomic.LoadUint64(&t.totalRequests),
		TotalMessages:   atomic.LoadUint64(&t.totalMessages),
		UselessRequests: atomic.LoadUint64(&t.uselessRequests),
		ActiveConns:     atomic.LoadInt64(&t.activeConns),
		MaxActiveConns:  atomic.LoadInt64(&t.maxActiveConns),
		MemMB:           currentMemMB(),
		TotalInc:        int64(t.incCount),
	}
}

func percentile(values []time.Duration, quantile float64) time.Duration {
	n := len(values)
	if n == 0 {
		return 0
	}
	index := int(float64(n-1) * quantile)
	if index < 0 {
		index = 0
	}
	if index >= n {
		index = n - 1
	}
	return values[index]
}

func currentMemMB() float64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return float64(stats.Alloc) / 1024 / 1024
}
