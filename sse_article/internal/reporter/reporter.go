package reporter

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

type Result struct {
	Protocol   string
	Clients    int
	P50        time.Duration
	P95        time.Duration
	P99        time.Duration
	Requests   int
	UselessPct float64
	MemMB      float64
}

func Print(result Result) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Protocolo\tClientes\tp50\tp95\tp99\tReq/min\tInútiles\tMem MB")
	fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%d\t%.1f%%\t%.1f\n",
		result.Protocol,
		result.Clients,
		result.P50,
		result.P95,
		result.P99,
		result.Requests,
		result.UselessPct,
		result.MemMB,
	)
	_ = w.Flush()
}
