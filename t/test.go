package main

import (
	"fmt"
	"github.com/lysu/metrics-logger"
	"time"
	//	"math/rand"
	"math/rand"
)

func main() {

	m := metrics.NewMonitor(20, 2*time.Second)
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Println(m.QPS(), "========")
		}
	}()
	go func() {
		for {
			t := rand.Intn(1000)
			time.Sleep(time.Duration(t) * time.Millisecond)
			for i := 0; i < 1000; i++ {
				m.RecordOne()
			}
		}
	}()
	time.Sleep(10 * time.Hour)

}
