package main

import (
	"context"
	"fmt"
	"log"
	"main/registry"
	"net/http"
)

func main() {
	http.Handle("/services", &registry.RegistryService{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var srv http.Server
	srv.Addr = registry.ServicePort

	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()

	go func() {
		log.Print("Service registry is running, press any key to shutdown")
		var s string
		fmt.Scan(&s)
		srv.Shutdown(ctx)
		cancel()
	}()
	<-ctx.Done()
	log.Print("Shutting down registry service")
}
