package engine

import "sync/atomic"

// Metrics tracks runtime counters for the proxy.
type Metrics struct {
	mutated    atomic.Int64
	dropped    atomic.Int64
	duplicated atomic.Int64
	delayed    atomic.Int64
}

// NewMetrics creates an empty metrics handle.
func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncMutated(delta int64) {
	if m != nil && delta != 0 {
		m.mutated.Add(delta)
	}
}

func (m *Metrics) IncDropped() {
	if m != nil {
		m.dropped.Add(1)
	}
}

func (m *Metrics) IncDuplicated(count int) {
	if m != nil && count > 0 {
		m.duplicated.Add(int64(count))
	}
}

func (m *Metrics) IncDelayed() {
	if m != nil {
		m.delayed.Add(1)
	}
}

// Snapshot returns the current counter values.
func (m *Metrics) Snapshot() map[string]int64 {
	if m == nil {
		return nil
	}
	return map[string]int64{
		"mutated":    m.mutated.Load(),
		"dropped":    m.dropped.Load(),
		"duplicated": m.duplicated.Load(),
		"delayed":    m.delayed.Load(),
	}
}
