package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type counter struct {
	c uint64
}

func (c *counter) increase() {
	atomic.AddUint64(&c.c, 1)
}

func (c *counter) reset() {
	atomic.StoreUint64(&c.c, 0)
}

type Monitor struct {
	startTime  time.Time
	epoch      int64
	windows    []counter
	windowSize int
	duration   time.Duration
	lock       sync.Locker
}

func NewMonitor(size int, duration time.Duration) *Monitor {
	start := time.Now()
	m := &Monitor{
		startTime:  start,
		epoch:      0,
		windows:    make([]counter, size),
		windowSize: size,
		duration:   duration,
		lock:       &sync.RWMutex{},
	}
	return m
}

func (m *Monitor) RecordOne() {
	current := time.Now()
	perWindowInterval := m.duration.Nanoseconds() / int64(m.windowSize)
	currentEpoch := current.Sub(m.startTime).Nanoseconds() / perWindowInterval
	m.lock.Lock()
	if currentEpoch > m.epoch {
		resetCount := 0
		for e := m.epoch + 1; e <= currentEpoch; e++ {
			resetIdx := int(e % int64(m.windowSize))
			m.windows[resetIdx].reset()
			resetCount++
			if resetCount >= m.windowSize {
				break
			}
		}
		m.epoch = currentEpoch
	}
	m.lock.Unlock()
	idx := int(currentEpoch % int64(m.windowSize))
	m.windows[idx].increase()
}

func (m *Monitor) QPS() float64 {
	var total uint64
	for _, counter := range m.windows {
		total += atomic.LoadUint64(&counter.c)
	}
	return float64(total) / m.duration.Seconds()
}
