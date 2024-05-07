package main

import (
	"context"
	"fmt"
	// stdlog "log"
	applog "main/services/app_log"
	"main/services/registry"
	"main/services/runner"
)

func main() {
	applog.Run("logs")

	host, port := "localhost", "4000"

	// serviceAddress := fmt.Sprintf("http://%v:%v", host, port)

	var r registry.Registration
	r.ServiceName = registry.LogService

	ctx, _ := runner.Start(context.Background(), host, port, r, applog.RegisterHandler)

	<-ctx.Done()
	fmt.Println("Shutting down LogService")
}
