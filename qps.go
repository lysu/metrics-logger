package metrics

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type bucket struct {
	count     uint64
	writeTime unsafe.Pointer
}

func (c *bucket) increase() {
	atomic.AddUint64(&c.count, 1)
	incTime := time.Now()
	atomic.StorePointer(&c.writeTime, unsafe.Pointer(&incTime))
}

func (c *bucket) reset() {
	atomic.StoreUint64(&c.count, 0)
}

// QPS is an sliding window QPS monitor
type QPS struct {
	startTime  time.Time
	epoch      int64
	windows    []bucket
	windowSize int
	duration   time.Duration
	lock       sync.Locker
}

// NewQPS uses to create new QPS monitor
func NewQPS(size int, duration time.Duration) *QPS {
	start := time.Now()
	m := &QPS{
		startTime:  start,
		epoch:      0,
		windows:    make([]bucket, size),
		windowSize: size,
		duration:   duration,
		lock:       &sync.RWMutex{},
	}
	return m
}

// RecordOne add one count into record
func (m *QPS) RecordOne() {
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

// QPS take current qps data.
func (m *QPS) QPS() float64 {
	current := time.Now()
	var total uint64
	for _, counter := range m.windows {
		pt := atomic.LoadPointer(&counter.writeTime)
		t := (*time.Time)(pt)
		if t != nil && current.Sub(*t) <= m.duration {
			total += atomic.LoadUint64(&counter.count)
		}
	}
	return float64(total) / m.duration.Seconds()
}
