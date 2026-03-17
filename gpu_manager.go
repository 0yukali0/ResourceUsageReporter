package main

import (
	"log"
	"sync"

	"nodeMonitor/model"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type GPUManager struct {
	mu          sync.Mutex
	initialized bool
}

func NewGPUManager() *GPUManager {
	manager := &GPUManager{}
	if manager.IsGPUNode() {
		return manager
	}
	return nil
}

// is_gpu_node 確認是否可以成功初始化 NVML。
// 成功代表此節點通常可視為 GPU node。
func (gm *GPUManager) IsGPUNode() bool {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.initialized {
		return true
	}

	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Printf("nvml init failed: %s", nvml.ErrorString(ret))
		return false
	}

	gm.initialized = true
	return true
}

// Close 用來釋放 NVML 資源。
func (gm *GPUManager) Close() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if !gm.initialized {
		return
	}

	ret := nvml.Shutdown()
	if ret != nvml.SUCCESS {
		log.Printf("nvml shutdown failed: %s", nvml.ErrorString(ret))
		return
	}

	gm.initialized = false
}

// Allocate 列出目前所有 GPU 資訊。
// 若非 GPU node 或 NVML 初始化失敗，回傳空 slice。
func (gm *GPUManager) Allocate() []model.GPUInfo {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if !gm.initialized {
		ret := nvml.Init()
		if ret != nvml.SUCCESS {
			log.Printf("nvml init failed: %s", nvml.ErrorString(ret))
			return []model.GPUInfo{}
		}
		gm.initialized = true
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Printf("get device count failed: %s", nvml.ErrorString(ret))
		return []model.GPUInfo{}
	}

	gpus := make([]model.GPUInfo, 0, count)

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Printf("get device handle by index %d failed: %s", i, nvml.ErrorString(ret))
			continue
		}

		name, ret := device.GetName()
		if ret != nvml.SUCCESS {
			log.Printf("get gpu name for index %d failed: %s", i, nvml.ErrorString(ret))
			name = "unknown"
		}

		util, ret := device.GetUtilizationRates()
		gpuUsage := 0.0
		if ret != nvml.SUCCESS {
			log.Printf("get utilization for index %d failed: %s", i, nvml.ErrorString(ret))
		} else {
			gpuUsage = float64(util.Gpu)
		}

		mem, ret := device.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			log.Printf("get memory info for index %d failed: %s", i, nvml.ErrorString(ret))
			continue
		}

		gpus = append(gpus, model.GPUInfo{
			Idx:      i,
			GPUName:  name,
			GPUUsage: gpuUsage,
			VRAM: model.VramInfo{
				Usage:     float64(mem.Used) / GB,
				Allocated: float64(mem.Used) / GB,
				Total:     float64(mem.Total) / GB,
			},
		})
	}

	return gpus
}
