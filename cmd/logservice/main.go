package main

import (
	"context"
	stlog "log"
	"main/log"
	"main/service"
)

func main() {
	log.Run("./app.log")
	host, port := "localhost", "4000"

	ctx, err := service.Start(context.Background(), "Log Service", host, port, log.RegisterHandlers)
	if err != nil {
		stlog.Fatal(err)
	}
	<-ctx.Done()
}
