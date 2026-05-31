// Package service contains domain logic for preparing metrics export payloads.
package service

import (
	"fmt"
	"runtime"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/domain/entity"
)

// CountersFromState builds counter metrics.
func CountersFromState(poll *entity.AgentPollMetrics) map[string]int64 {
	return map[string]int64{
		"PollCount": poll.PollCount,
	}
}

// GaugesFromState builds gauge metric values from runtime, poll, and host stats.
func GaugesFromState(ms *runtime.MemStats, poll *entity.AgentPollMetrics, gs *entity.GopsutilMetrics) map[string]float64 {
	m := map[string]float64{
		"Alloc":         float64(ms.Alloc),
		"BuckHashSys":   float64(ms.BuckHashSys),
		"Frees":         float64(ms.Frees),
		"GCCPUFraction": ms.GCCPUFraction,
		"GCSys":         float64(ms.GCSys),
		"HeapAlloc":     float64(ms.HeapAlloc),
		"HeapIdle":      float64(ms.HeapIdle),
		"HeapInuse":     float64(ms.HeapInuse),
		"HeapObjects":   float64(ms.HeapObjects),
		"HeapReleased":  float64(ms.HeapReleased),
		"HeapSys":       float64(ms.HeapSys),
		"LastGC":        float64(ms.LastGC),
		"Lookups":       float64(ms.Lookups),
		"MCacheInuse":   float64(ms.MCacheInuse),
		"MCacheSys":     float64(ms.MCacheSys),
		"MSpanInuse":    float64(ms.MSpanInuse),
		"MSpanSys":      float64(ms.MSpanSys),
		"Mallocs":       float64(ms.Mallocs),
		"NextGC":        float64(ms.NextGC),
		"NumForcedGC":   float64(ms.NumForcedGC),
		"NumGC":         float64(ms.NumGC),
		"OtherSys":      float64(ms.OtherSys),
		"PauseTotalNs":  float64(ms.PauseTotalNs),
		"StackInuse":    float64(ms.StackInuse),
		"StackSys":      float64(ms.StackSys),
		"Sys":           float64(ms.Sys),
		"TotalAlloc":    float64(ms.TotalAlloc),
		"RandomValue":   poll.RandomValue,
		"TotalMemory":   gs.TotalMemory,
		"FreeMemory":    gs.FreeMemory,
	}
	for i, cpuUtil := range gs.CPUutilization {
		m[fmt.Sprintf("CPUutilization%d", i+1)] = cpuUtil
	}
	return m
}
