package metrics

import (
	"runtime"
	"time"

	"github.com/rcrowley/go-metrics"
)

// ReportMemstatsMetrics will send runtime Memstats metrics every 10 seconds,
// and will block forever.
func ReportMemstatsMetrics() {
	memStats := &runtime.MemStats{}
	lastSampleTime := time.Now()
	var lastPauseNs uint64
	var lastNumGC uint64

	sleep := 10 * time.Second

	for {
		runtime.ReadMemStats(memStats)

		now := time.Now()

		metrics.GetOrRegisterGauge("travis.gcloud-cleanup.goroutines", metrics.DefaultRegistry).Update(int64(runtime.NumGoroutine()))
		metrics.GetOrRegisterGauge("travis.gcloud-cleanup.memory.allocated", metrics.DefaultRegistry).Update(int64(memStats.Alloc))
		metrics.GetOrRegisterGauge("travis.gcloud-cleanup.memory.mallocs", metrics.DefaultRegistry).Update(int64(memStats.Mallocs))
		metrics.GetOrRegisterGauge("travis.gcloud-cleanup.memory.frees", metrics.DefaultRegistry).Update(int64(memStats.Frees))
		metrics.GetOrRegisterGauge("travis.gcloud-cleanup.memory.gc.total_pause", metrics.DefaultRegistry).Update(int64(memStats.PauseTotalNs))
		metrics.GetOrRegisterGauge("travis.gcloud-cleanup.memory.gc.heap", metrics.DefaultRegistry).Update(int64(memStats.HeapAlloc))
		metrics.GetOrRegisterGauge("travis.gcloud-cleanup.memory.gc.stack", metrics.DefaultRegistry).Update(int64(memStats.StackInuse))

		if lastPauseNs > 0 {
			pauseSinceLastSample := memStats.PauseTotalNs - lastPauseNs
			metrics.GetOrRegisterGauge("travis.gcloud-cleanup.memory.gc.pause_per_second", metrics.DefaultRegistry).Update(int64(float64(pauseSinceLastSample) / sleep.Seconds()))
		}
		lastPauseNs = memStats.PauseTotalNs

		countGC := int(uint64(memStats.NumGC) - lastNumGC)
		if lastNumGC > 0 {
			diff := float64(countGC)
			diffTime := now.Sub(lastSampleTime).Seconds()
			metrics.GetOrRegisterGauge("travis.gcloud-cleanup.memory.gc.gc_per_second", metrics.DefaultRegistry).Update(int64(diff / diffTime))
		}

		if countGC > 0 {
			if countGC > 256 {
				countGC = 256
			}

			for i := 0; i < countGC; i++ {
				idx := int((memStats.NumGC-uint32(i))+255) % 256
				pause := time.Duration(memStats.PauseNs[idx])
				metrics.GetOrRegisterTimer("travis.gcloud-cleanup.memory.gc.pause", metrics.DefaultRegistry).Update(pause)
			}
		}

		lastNumGC = uint64(memStats.NumGC)
		lastSampleTime = now

		time.Sleep(sleep)
	}
}
