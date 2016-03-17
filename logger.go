package metrics

import (
	"runtime"
	"time"

	"github.com/golang/glog"
)

const (
	GoroutineCount   = "GoroutineCount"
	MemoryAllocated  = "MemoryAllocated"
	MemoryMallocs    = "MemoryMallocs"
	MemoryFrees      = "MemoryFrees"
	MemoryHeap       = "MemoryHeap"
	MemoryStack      = "MemoryStack"
	GcPauseTime      = "GcPauseTime"
	GcTotalPause     = "GcTotalPause"
	GcPausePerSecond = "GcPausePerSecond"
	GcPerSecond      = "GcPerSecond"
)

func Time(key string, startTime time.Time, timeThreshold time.Duration) {
	timeSpent := time.Now().Sub(startTime)
	if timeSpent > timeThreshold {
		glog.Infof("[GoMetric]time spent: %d, for %s", timeSpent.Nanoseconds()/int64(time.Millisecond), key)
	}
}

func CountOne(key string) {
	Count(key, 1)
}

func Count(key string, num int) {
	glog.Infof("[GoMetric]counter increase %d, for %s", num, key)
}

func Gauga(key string, value float64) {
	glog.Infof("[GoMetric]gauga set %2f, for %s", value, key)
}

func GoMonitor(interval time.Duration) {
	go func() {

		defer func() {
			if r := recover(); r != nil {
				glog.Warningf("Monitor goroutine is panic %v \n", r)
			}
		}()

		memStats := &runtime.MemStats{}
		lastSampleTime := time.Now()
		var lastPauseNs uint64 = 0
		var lastNumGc uint32 = 0

		nsInMs := float64(time.Millisecond)

		for range time.Tick(interval) {

			runtime.ReadMemStats(memStats)

			now := time.Now()

			Gauga(GoroutineCount, float64(runtime.NumGoroutine()))
			Gauga(MemoryAllocated, float64(memStats.Alloc))
			Gauga(MemoryMallocs, float64(memStats.Mallocs))
			Gauga(MemoryFrees, float64(memStats.Frees))
			Gauga(MemoryHeap, float64(memStats.HeapAlloc))
			Gauga(MemoryStack, float64(memStats.StackInuse))
			Gauga(GcTotalPause, float64(memStats.PauseTotalNs)/nsInMs)

			if lastPauseNs > 0 {
				pauseSinceLastSample := memStats.PauseTotalNs - lastPauseNs
				Gauga(GcPausePerSecond, float64(pauseSinceLastSample)/nsInMs/interval.Seconds())
			}
			lastPauseNs = memStats.PauseTotalNs

			countGc := int(memStats.NumGC - lastNumGc)
			if lastNumGc > 0 {
				diff := float64(countGc)
				diffTime := now.Sub(lastSampleTime).Seconds()
				Gauga(GcPerSecond, float64(diff)/diffTime)
			}

			if countGc > 0 {
				if countGc > 256 {
					countGc = 256
				}
				for i := 0; i < countGc; i++ {
					idx := int((memStats.NumGC-uint32(i))+255) % 256
					pause := float64(memStats.PauseNs[idx])
					Gauga(GcPauseTime, float64(pause/nsInMs))
				}
			}

			lastNumGc = memStats.NumGC
			lastSampleTime = now

		}
	}()
}
