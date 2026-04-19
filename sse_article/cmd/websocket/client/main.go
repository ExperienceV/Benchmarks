package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"benchmark/internal/metrics"
	"benchmark/internal/reporter"

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
	var clients int
	var duration time.Duration
	var host string
	var port int
	flag.IntVar(&clients, "clients", 10, "number of concurrent clients")
	flag.DurationVar(&duration, "duration", 60*time.Second, "measurement duration")
	flag.StringVar(&host, "host", "127.0.0.1", "server host")
	flag.IntVar(&port, "port", 8081, "server port")
	flag.Parse()

	tracker := metrics.New()
	ctx, cancel := context.WithTimeout(context.Background(), duration+10*time.Second)
	defer cancel()

	measurementCtx, measurementCancel := context.WithTimeout(ctx, 5*time.Second+duration)
	defer measurementCancel()

	startTime := time.Now().Add(5 * time.Second)
	wg := sync.WaitGroup{}
	wg.Add(clients)
	for i := 0; i < clients; i++ {
		go clientWorker(measurementCtx, tracker, fmt.Sprintf("ws://%s:%d/ws", host, port), startTime, &wg)
	}
	wg.Wait()

	if err := report(host, port, duration, tracker, clients); err != nil {
		log.Printf("report error: %v", err)
	}
}

func clientWorker(ctx context.Context, tracker *metrics.Tracker, endpoint string, warmUpEnd time.Time, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, _, err := websocket.Dial(ctx, endpoint, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	for {
		var msg message
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return
		}
		if time.Now().Before(warmUpEnd) {
			continue
		}
		tracker.AddMessage()
		tracker.RecordLatency(time.Since(msg.Timestamp))
		tracker.AddRequest()
	}
}

func report(host string, port int, duration time.Duration, tracker *metrics.Tracker, clients int) error {
	snapshot := tracker.Snapshot()
	serverMetricsURL := fmt.Sprintf("http://%s:%d/metrics", host, port)
	var m metricsResponse
	if err := getJSON(serverMetricsURL, &m); err != nil {
		return err
	}

	reporter.Print(reporter.Result{
		Protocol:   "WebSocket",
		Clients:    clients,
		P50:        snapshot.P50,
		P95:        snapshot.P95,
		P99:        snapshot.P99,
		Requests:   int(float64(snapshot.TotalMessages) * 60 / duration.Seconds()),
		UselessPct: 0,
		MemMB:      m.MemMB,
	})
	return nil
}

func getJSON(endpoint string, dest interface{}) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dest)
}
