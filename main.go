package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

func handleConsole(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data, err := os.ReadFile("console.html")
	if err != nil {
		log.Printf("failed to read console.html: %v", err)
		http.Error(w, "Console not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func main() {
	server := NewServer()
	gpu_manager := NewGPUManager()

	router := httprouter.New()
	router.GET("/console", handleConsole)
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
			res, currentBytes, err := collectResource(gpu_manager, prevBytes, 1.0)
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
