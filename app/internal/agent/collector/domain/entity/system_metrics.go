// Package entity holds in-memory system metrics.
package entity

import "runtime"

// RuntimeMetrics holds process runtime memory statistics.
type RuntimeMetrics struct {
	MemStats runtime.MemStats
}

// AgentPollMetrics holds values updated on each agent poll tick.
type AgentPollMetrics struct {
	PollCount   int64
	RandomValue float64
}

// GopsutilMetrics holds memory and per-CPU utilization from host sampling.
type GopsutilMetrics struct {
	TotalMemory    float64
	FreeMemory     float64
	CPUutilization []float64
}

// SystemMetrics is a point-in-time snapshot for export (runtime, host, poll-derived).
type SystemMetrics struct {
	Runtime  RuntimeMetrics
	Poll     AgentPollMetrics
	Gopsutil GopsutilMetrics
}
