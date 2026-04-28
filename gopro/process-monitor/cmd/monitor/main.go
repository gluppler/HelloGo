package main

import (
	"context"
	"log"
	"time"

	"github.com/gluppler/process-monitor/internal/process"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dashboard := process.NewDashboard(ctx, 2*time.Second)
	defer dashboard.Stop()

	if err := dashboard.Refresh(); err != nil {
		log.Fatalf("failed to refresh: %v", err)
	}

	print(dashboard.Render())
}