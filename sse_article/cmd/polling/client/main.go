package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"benchmark/internal/metrics"
	"benchmark/internal/reporter"
)

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
	var clients int
	var duration time.Duration
	var host string
	var port int
	flag.IntVar(&clients, "clients", 10, "number of concurrent clients")
	flag.DurationVar(&duration, "duration", 60*time.Second, "measurement duration")
	flag.StringVar(&host, "host", "127.0.0.1", "server host")
	flag.IntVar(&port, "port", 8080, "server port")
	flag.Parse()

	tracker := metrics.New()
	ctx, cancel := context.WithTimeout(context.Background(), duration+10*time.Second)
	defer cancel()

	warmUpEnd := time.Now().Add(5 * time.Second)
	measurementCtx, measurementCancel := context.WithDeadline(ctx, warmUpEnd.Add(duration))
	defer measurementCancel()

	wg := sync.WaitGroup{}
	wg.Add(clients)
	for i := 0; i < clients; i++ {
		go worker(measurementCtx, tracker, fmt.Sprintf("http://%s:%d/latest", host, port), warmUpEnd, &wg)
	}

	wg.Wait()
	if err := report(host, port, duration, tracker, clients); err != nil {
		log.Printf("report error: %v", err)
	}
}

func worker(ctx context.Context, tracker *metrics.Tracker, endpoint string, warmUpEnd time.Time, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{Timeout: 10 * time.Second}
	var lastVersion int64
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reqURL, _ := url.Parse(endpoint)
			q := reqURL.Query()
			q.Set("since", fmt.Sprint(lastVersion))
			reqURL.RawQuery = q.Encode()

			resp, err := client.Get(reqURL.String())
			if err != nil {
				continue
			}
			var payload response
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				_ = resp.Body.Close()
				continue
			}
			_ = resp.Body.Close()

			if time.Now().Before(warmUpEnd) {
				continue
			}

			tracker.AddRequest()
			if payload.Changed {
				lastVersion = payload.Version
				tracker.RecordLatency(time.Since(payload.Timestamp))
				tracker.AddMessage()
			} else {
				tracker.AddUselessRequest()
			}
		}
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
		Protocol:   "Polling",
		Clients:    clients,
		P50:        snapshot.P50,
		P95:        snapshot.P95,
		P99:        snapshot.P99,
		Requests:   int(float64(snapshot.TotalRequests) * 60 / duration.Seconds()),
		UselessPct: percent(snapshot.UselessRequests, snapshot.TotalRequests),
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

func percent(part, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
