package model

import (
	"time"
)

type Memory struct {
	Usage           float64 `json:"usage"`
	AllocatedMemory float64 `json:"allocated_memory"`
	TotalMemory     float64 `json:"total_memory"`
}

type Disk struct {
	Usage float64 `json:"usage"`
	Total float64 `json:"total"`
}

type VramInfo struct {
	Usage     float64 `json:"usage"`
	Allocated float64 `json:"allocated"`
	Total     float64 `json:"total"`
}

type GPUInfo struct {
	Idx      int      `json:"idx"`
	GPUName  string   `json:"gpu_name"`
	GPUUsage float64  `json:"gpu_usage"`
	VRAM     VramInfo `json:"vram"`
}

type Resource struct {
	Timestamp time.Time `json:"timestamp"`
	Node      string    `json:"node"`
	IP        string    `json:"ip"`
	CPU       float64   `json:"cpu"`
	Memory    Memory    `json:"memory"`
	Bandwidth float64   `json:"bandwitdh"`
	Disk      Disk      `json:"disk"`
	GPUs      []GPUInfo `json:"gpus,omitempty"`
}
