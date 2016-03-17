package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type counter struct {
	c *uint64
}

func newCounter() counter {
	cc := uint64(0)
	return counter{
		c: &cc,
	}
}

func (c *counter) increase() {
	atomic.AddUint64(c.c, 1)
}

func (c *counter) reset() {
	atomic.StoreUint64(c.c, 0)
}

type Metrics struct {
	startTime  time.Time
	currentIdx int
	windows    []counter
	windowSize int
	duration   time.Duration
	lock       sync.Locker
}

func NewMetrics(size int, duration time.Duration) *Metrics {
	start := time.Now()
	m := &Metrics{
		startTime:  start,
		currentIdx: 0,
		windows:    make([]counter, 0, size),
		windowSize: size,
		duration:   duration,
		lock:       &sync.RWMutex{},
	}
	for i := 0; i < size; i++ {
		m.windows = append(m.windows, newCounter())
	}
	return m
}

func (m *Metrics) RecordOne() {
	current := time.Now()

	perWindowNano := m.duration.Nanoseconds() / int64(m.windowSize)

	idx := int(current.Sub(m.startTime).Nanoseconds() / perWindowNano % int64(m.windowSize))
	m.lock.Lock()
	if m.currentIdx != idx {
		m.currentIdx = idx
		m.windows[idx].reset()
	}
	m.lock.Unlock()
	m.windows[idx].increase()
}

func (m *Metrics) QPS() float64 {
	var total uint64
	for _, counter := range m.windows {
		total += atomic.LoadUint64(counter.c)

	}
	return float64(total) / m.duration.Seconds()
}
