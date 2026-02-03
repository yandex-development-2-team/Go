package main

import (
	"context"

	".../internal/shutdown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sh := shutdown.NewShutdownHandler()

	go func() {
		if err := sh.WaitForShutdown(ctx, cancel,
			shutdown.ShutdownTask{
				Name: "TGUpdates",
				Fn:   updates.Stop,
			},
			shutdown.ShutdownTask{
				Name: "Database",
				Fn:   db.Close,
			},
			shutdown.ShutdownTask{
				Name: "Prometheus metrics",
				Fn:   metricsServer.Stop,
			},
		); err != nil {
			//logger
		}
	}()
	<-ctx.Done()
}
