package metrics

import (
	"runtime"
	"time"
)

type Metrics struct {
	PollCount     int64
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64
	RandomValue   float64
}

func NewMetrics() *Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &Metrics{
		Alloc:         float64(memStats.Alloc),
		BuckHashSys:   float64(memStats.BuckHashSys),
		Frees:         float64(memStats.Frees),
		GCCPUFraction: memStats.GCCPUFraction,
		GCSys:         float64(memStats.GCSys),
		HeapAlloc:     float64(memStats.HeapAlloc),
		HeapIdle:      float64(memStats.HeapIdle),
		HeapInuse:     float64(memStats.HeapInuse),
		HeapObjects:   float64(memStats.HeapObjects),
		HeapReleased:  float64(memStats.HeapReleased),
		HeapSys:       float64(memStats.HeapSys),
		LastGC:        float64(memStats.LastGC),
		Lookups:       float64(memStats.Lookups),
		MCacheInuse:   float64(memStats.MCacheInuse),
		MCacheSys:     float64(memStats.MCacheSys),
		MSpanInuse:    float64(memStats.MSpanInuse),
		MSpanSys:      float64(memStats.MSpanSys),
		Mallocs:       float64(memStats.Mallocs),
		NextGC:        float64(memStats.NextGC),
		NumForcedGC:   float64(memStats.NumForcedGC),
		NumGC:         float64(memStats.NumGC),
		OtherSys:      float64(memStats.OtherSys),
		PauseTotalNs:  float64(memStats.PauseTotalNs),
		StackInuse:    float64(memStats.StackInuse),
		StackSys:      float64(memStats.StackSys),
		Sys:           float64(memStats.Sys),
		TotalAlloc:    float64(memStats.TotalAlloc),
		RandomValue:   float64(time.Now().Unix()),
	}
}
