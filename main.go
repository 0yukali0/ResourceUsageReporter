package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"
)

const (
	GB = 1024 * 1024 * 1024
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

type Server struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func NewServer() *Server {

	return &Server{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (s *Server) addClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[conn] = struct{}{}
}

func (s *Server) removeClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, conn)
	_ = conn.Close()
}

func (s *Server) broadcast(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("write message error: %v", err)
			_ = conn.Close()
			delete(s.clients, conn)
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// demo 用，正式環境請限制 origin
		return true
	},
}

func (s *Server) handleResourcesWS(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade websocket failed: %v", err)
		return
	}

	s.addClient(conn)
	log.Printf("client connected: %s", conn.RemoteAddr())

	go func() {
		defer func() {
			s.removeClient(conn)
			log.Printf("client disconnected: %s", conn.RemoteAddr())
		}()

		// 持續 read 是為了能偵測 client 關閉
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

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

func collectResource(prevBytes uint64, intervalSec float64) (Resource, uint64, error) {
	now := time.Now()

	host := getHostname()
	ip := getLocalIP()

	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return Resource{}, prevBytes, err
	}
	cpuUsage := 0.0
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return Resource{}, prevBytes, err
	}

	du, err := disk.Usage(getDiskPath())
	if err != nil {
		return Resource{}, prevBytes, err
	}

	currentBytes, err := getTotalNetworkBytes()
	if err != nil {
		return Resource{}, prevBytes, err
	}

	var bandwidth float64
	if prevBytes > 0 && currentBytes >= prevBytes && intervalSec > 0 {
		// bytes/s -> bit/s
		bandwidth = float64(currentBytes-prevBytes) * 8 / intervalSec
	}

	gpu_manager := NewGPUManager()

	res := Resource{
		Timestamp: now,
		Node:      host,
		IP:        ip,
		CPU:       cpuUsage,
		Memory: Memory{
			Usage:           float64(vm.Used) / GB,
			AllocatedMemory: float64(vm.Total-vm.Available) / GB,
			TotalMemory:     float64(vm.Total) / GB,
		},
		Bandwidth: bandwidth,
		Disk: Disk{
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

func main() {
	server := NewServer()

	router := httprouter.New()
	router.GET("/resources", server.handleResourcesWS)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("server listening on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		prevBytes, err := getTotalNetworkBytes()
		if err != nil {
			log.Printf("initial network bytes error: %v", err)
		}

		for range ticker.C {
			res, currentBytes, err := collectResource(prevBytes, 1.0)
			if err != nil {
				log.Printf("collect resource error: %v", err)
				continue
			}

			prevBytes = currentBytes
			server.broadcast(res)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
