./run_benchmarks.sh 500 60s   # Quick Test (1 min)

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

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type message struct {
	Timestamp time.Time `json:"timestamp"`
	Sequence  int64     `json:"sequence"`
}

type metricsResponse struct {
	ActiveConns int64   `json:"active_conns"`
	MemMB       float64 `json:"mem_mb"`
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "server port")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tracker := metrics.New()
	hub := newHub(tracker)
	go hub.run(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			http.Error(w, "upgrade failed", http.StatusBadRequest)
			return
		}
		hub.add(conn)
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		snapshot := tracker.Snapshot()
		mr := metricsResponse{ActiveConns: snapshot.MaxActiveConns, MemMB: snapshot.MemMB}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mr)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		log.Printf("websocket server listening on %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

type hub struct {
	mu       sync.Mutex
	conns    map[*websocket.Conn]struct{}
	tracker  *metrics.Tracker
	sequence int64
}

func newHub(tracker *metrics.Tracker) *hub {
	return &hub{conns: make(map[*websocket.Conn]struct{}), tracker: tracker}
}

func (h *hub) add(conn *websocket.Conn) {
	h.mu.Lock()
	h.conns[conn] = struct{}{}
	h.mu.Unlock()
	h.tracker.IncActive()

	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.conns, conn)
			h.mu.Unlock()
			h.tracker.DecActive()
			_ = conn.Close(websocket.StatusNormalClosure, "client disconnected")
		}()
		for {
			_, _, err := conn.Read(context.Background())
			if err != nil {
				return
			}
		}
	}()
}

func (h *hub) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			h.mu.Lock()
			h.sequence++
			msg := message{Timestamp: now.UTC(), Sequence: h.sequence}
			for conn := range h.conns {
				_ = wsjson.Write(context.Background(), conn, msg)
			}
			h.mu.Unlock()
		}
	}
}
