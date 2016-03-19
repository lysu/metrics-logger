package metrics_test

import (
	"github.com/lysu/metrics-logger"
	"testing"
	"time"
)

func BenchmarkQPS(b *testing.B) {
	m := metrics.NewQPS(20, 2*time.Second)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RecordOne()
		m.QPS()
	}
}
