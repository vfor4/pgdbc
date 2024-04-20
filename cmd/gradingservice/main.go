package main

import (
	"context"
	"fmt"
	stdlog "log"
	"main/grade"
	"main/registry"
	"main/service"
)

func main() {
	host, port := "localhost", "6000"
	serviceAddress := fmt.Sprintf("http://%v:%v", host, port)

	var r registry.Registration

	r.ServiceName = registry.GradingService
	r.ServiceUrl = serviceAddress

	ctx, err := service.Start(context.Background(), host, port, r, grade.RegisterHandlers)
	if err != nil {
		stdlog.Fatal(err)
	}
	<-ctx.Done()
	fmt.Print("Shutting down grading server")
}
