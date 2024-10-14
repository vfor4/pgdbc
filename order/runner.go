package order

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func Start(ctx context.Context, host, port string) (context.Context, error) {
	RegisterHandlers()
	ctx = startService(ctx, host, port) //refactor the error
	return ctx, nil
}

func startService(ctx context.Context, host, port string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	var srv http.Server
	srv.Addr = fmt.Sprintf("%s:%s", host, port)
	go func() {
		fmt.Print(srv.ListenAndServe())
		cancel()
	}()

	go func() {
		fmt.Println("service started, press any keys to shutdown")
		var s string
		fmt.Scanln(&s)
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Print(err)
		}
		cancel()

	}()
	return ctx
}
