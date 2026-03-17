package main

import (
	"net"
	"nodeMonitor/model"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"
)

const (
	GB = 1024 * 1024 * 1024
)

func getHostname() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "unknown"
	}
	return name
}

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip4 := ip.To4()
			if ip4 != nil {
				return ip4.String()
			}
		}
	}

	return "127.0.0.1"
}

func getDiskPath() string {
	if os.PathSeparator == '\\' {
		return "C:\\"
	}
	return "/"
}

func getTotalNetworkBytes() (uint64, error) {
	counters, err := gnet.IOCounters(false)
	if err != nil {
		return 0, err
	}
	if len(counters) == 0 {
		return 0, nil
	}

	total := counters[0].BytesRecv + counters[0].BytesSent
	return total, nil
}

func collectResource(prevBytes uint64, intervalSec float64) (model.Resource, uint64, error) {
	now := time.Now()

	host := getHostname()
	ip := getLocalIP()

	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return model.Resource{}, prevBytes, err
	}
	cpuUsage := 0.0
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return model.Resource{}, prevBytes, err
	}

	du, err := disk.Usage(getDiskPath())
	if err != nil {
		return model.Resource{}, prevBytes, err
	}

	currentBytes, err := getTotalNetworkBytes()
	if err != nil {
		return model.Resource{}, prevBytes, err
	}

	var bandwidth float64
	if prevBytes > 0 && currentBytes >= prevBytes && intervalSec > 0 {
		// bytes/s -> bit/s
		bandwidth = float64(currentBytes-prevBytes) * 8 / intervalSec
	}

	gpu_manager := NewGPUManager()

	res := model.Resource{
		Timestamp: now,
		Node:      host,
		IP:        ip,
		CPU:       cpuUsage,
		Memory: model.Memory{
			Usage:           float64(vm.Used) / GB,
			AllocatedMemory: float64(vm.Total-vm.Available) / GB,
			TotalMemory:     float64(vm.Total) / GB,
		},
		Bandwidth: bandwidth,
		Disk: model.Disk{
			Usage: float64(du.Used) / GB,
			Total: float64(du.Total) / GB,
		},
		GPUs: nil,
	}

	if gpu_manager != nil {
		res.GPUs = gpu_manager.Allocate()
	}

	return res, currentBytes, nil
}
