package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"benchmark/internal/metrics"
	"benchmark/internal/reporter"
)

type eventPayload struct {
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
	flag.IntVar(&port, "port", 8082, "server port")
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
		go streamClient(measurementCtx, tracker, fmt.Sprintf("http://%s:%d/events", host, port), startTime, &wg)
	}
	wg.Wait()

	if err := report(host, port, duration, tracker, clients); err != nil {
		log.Printf("report error: %v", err)
	}
}

func streamClient(ctx context.Context, tracker *metrics.Tracker, endpoint string, warmUpEnd time.Time, wg *sync.WaitGroup) {
	defer wg.Done()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var builder strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			eventText := builder.String()
			builder.Reset()
			if eventText != "" {
				processEvent(eventText, tracker, warmUpEnd)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "data:") {
			builder.WriteString(strings.TrimSpace(strings.TrimPrefix(trimmed, "data:")))
		}
	}
}

func processEvent(payload string, tracker *metrics.Tracker, warmUpEnd time.Time) {
	var ev eventPayload
	if err := json.Unmarshal([]byte(payload), &ev); err != nil {
		return
	}
	if time.Now().Before(warmUpEnd) {
		return
	}
	tracker.AddMessage()
	tracker.RecordLatency(time.Since(ev.Timestamp))
	tracker.AddRequest()
}

func report(host string, port int, duration time.Duration, tracker *metrics.Tracker, clients int) error {
	snapshot := tracker.Snapshot()
	serverMetricsURL := fmt.Sprintf("http://%s:%d/metrics", host, port)
	var m metricsResponse
	if err := getJSON(serverMetricsURL, &m); err != nil {
		return err
	}

	reporter.Print(reporter.Result{
		Protocol:   "SSE",
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
