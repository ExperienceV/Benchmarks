package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"benchmark/internal/metrics"
)

type state struct {
	mu        sync.RWMutex
	timestamp time.Time
	version   int64
}

type response struct {
	Changed   bool      `json:"changed"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Version   int64     `json:"version,omitempty"`
}

type metricsResponse struct {
	ActiveConns int64   `json:"active_conns"`
	MemMB       float64 `json:"mem_mb"`
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "server port")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tracker := metrics.New()
	s := &state{timestamp: time.Now(), version: 1}

	go ticker(ctx, s)

	mux := http.NewServeMux()
	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		tracker.IncActive()
		defer tracker.DecActive()
		tracker.AddRequest()

		since := int64(0)
		if q := r.URL.Query().Get("since"); q != "" {
			fmt.Sscan(q, &since)
		}

		s.mu.RLock()
		changed := s.version > since
		resp := response{Changed: changed}
		if changed {
			resp.Timestamp = s.timestamp
			resp.Version = s.version
		}
		s.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		snapshot := tracker.Snapshot()
		mr := metricsResponse{ActiveConns: snapshot.TotalInc, MemMB: snapshot.MemMB}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mr)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		log.Printf("polling server listening on %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func ticker(ctx context.Context, s *state) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.mu.Lock()
			s.timestamp = now.UTC()
			s.version++
			s.mu.Unlock()
		}
	}
}
