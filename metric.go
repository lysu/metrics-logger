package metrics

import (
	"runtime"
	"time"

	"github.com/golang/glog"
)

const SYSTEM_LOG_ID = "GO"

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

func Time(logid string, key string, startTime time.Time, timeThreshold time.Duration) {
	timeSpent := time.Now().Sub(startTime)
	if timeSpent > timeThreshold {
		glog.Infof("%s, time spent: %d, for %s", logid, timeSpent.Nanoseconds()/int64(time.Millisecond), key)
	}
}

func CountOne(logid string, key string) {
	Count(logid, key, 1)
}

func Count(logid string, key string, num int) {
	glog.Infof("%s, counter increase %d, for %s", logid, num, key)
}

func Gauga(logid string, key string, value float64) {
	glog.Infof("%s, gauga set %2f, for %s", logid, value, key)
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

			Gauga(SYSTEM_LOG_ID, GoroutineCount, float64(runtime.NumGoroutine()))
			Gauga(SYSTEM_LOG_ID, MemoryAllocated, float64(memStats.Alloc))
			Gauga(SYSTEM_LOG_ID, MemoryMallocs, float64(memStats.Mallocs))
			Gauga(SYSTEM_LOG_ID, MemoryFrees, float64(memStats.Frees))
			Gauga(SYSTEM_LOG_ID, MemoryHeap, float64(memStats.HeapAlloc))
			Gauga(SYSTEM_LOG_ID, MemoryStack, float64(memStats.StackInuse))
			Gauga(SYSTEM_LOG_ID, GcTotalPause, float64(memStats.PauseTotalNs)/nsInMs)

			if lastPauseNs > 0 {
				pauseSinceLastSample := memStats.PauseTotalNs - lastPauseNs
				Gauga(SYSTEM_LOG_ID, GcPausePerSecond, float64(pauseSinceLastSample)/nsInMs/interval.Seconds())
			}
			lastPauseNs = memStats.PauseTotalNs

			countGc := int(memStats.NumGC - lastNumGc)
			if lastNumGc > 0 {
				diff := float64(countGc)
				diffTime := now.Sub(lastSampleTime).Seconds()
				Gauga(SYSTEM_LOG_ID, GcPerSecond, float64(diff)/diffTime)
			}

			if countGc > 0 {
				if countGc > 256 {
					countGc = 256
				}
				for i := 0; i < countGc; i++ {
					idx := int((memStats.NumGC-uint32(i))+255) % 256
					pause := float64(memStats.PauseNs[idx])
					Gauga(SYSTEM_LOG_ID, GcPauseTime, float64(pause/nsInMs))
				}
			}

			lastNumGc = memStats.NumGC
			lastSampleTime = now

		}
	}()
}
