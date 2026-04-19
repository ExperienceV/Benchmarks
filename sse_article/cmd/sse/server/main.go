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

type metricsResponse struct {
	ActiveConns int64   `json:"active_conns"`
	MemMB       float64 `json:"mem_mb"`
}

type subscriber struct {
	events chan string
	done   chan struct{}
}

type broker struct {
	mu      sync.Mutex
	subs    map[*subscriber]struct{}
	tracker *metrics.Tracker
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "server port")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tracker := metrics.New()
	b := &broker{subs: make(map[*subscriber]struct{}), tracker: tracker}
	go b.run(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/events", b.handleEvents)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		snapshot := tracker.Snapshot()
		mr := metricsResponse{ActiveConns: snapshot.MaxActiveConns, MemMB: snapshot.MemMB}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mr)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		log.Printf("sse server listening on %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func (b *broker) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	sub := &subscriber{events: make(chan string, 10), done: make(chan struct{})}
	b.mu.Lock()
	b.subs[sub] = struct{}{}
	b.mu.Unlock()
	b.tracker.IncActive()
	defer func() {
		b.mu.Lock()
		delete(b.subs, sub)
		b.mu.Unlock()
		b.tracker.DecActive()
		close(sub.done)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()

	for {
		select {
		case msg := <-sub.events:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ctx.Done():
			return
		case <-sub.done:
			return
		}
	}
}

func (b *broker) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	var sequence int64
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			sequence++
			event := fmt.Sprintf(`{"timestamp":"%s","sequence":%d}`, now.UTC().Format(time.RFC3339Nano), sequence)
			b.mu.Lock()
			for sub := range b.subs {
				select {
				case sub.events <- event:
				default:
				}
			}
			b.mu.Unlock()
		}
	}
}
