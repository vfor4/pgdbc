package main

import (
	"context"
	"fmt"
	"log"
	"main/services/registry"
	"net/http"
)

func main() {
	http.Handle("/services", &registry.RegistryService{})

	ctx, cancel := context.WithCancel(context.Background())

	var srv http.Server
	srv.Addr = ":3000"

	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()

	go func() {
		fmt.Println("Registry service started. Press any key to stop ")
		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done()
	log.Println("Shutting down Registry service")
}
