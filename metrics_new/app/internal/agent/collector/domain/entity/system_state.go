package entity

import "runtime"

// AgentPollStats holds values updated on each agent poll tick.
type AgentPollStats struct {
	PollCount   int64
	RandomValue float64
}

// GopsutilStats holds memory and per-CPU utilization from host sampling.
type GopsutilStats struct {
	TotalMemory    float64
	FreeMemory     float64
	CPUutilization []float64
}

// SystemState is a point-in-time snapshot of runtime memstats, host sample, and poll-derived fields.
type SystemState struct {
	MemStats runtime.MemStats
	Poll     AgentPollStats
	Gopsutil GopsutilStats
}
