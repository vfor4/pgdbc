package runner

import (
	"context"
	"fmt"
	"log"
	stdlog "log"
	"main/services/registry"
	"net/http"
)

func Start(ctx context.Context, host, port string, r registry.Registration, registerHandlersFunc func()) (context.Context, error) {
	registerHandlersFunc()
	ctx = startService(ctx, r.ServiceName, host, port)
	err := registry.RegisterService(r)
	if err != nil {
		log.Println(err)
		return ctx, err
	}
	return ctx, nil
}

func startService(ctx context.Context, serviceName registry.ServiceName, host, port string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	var srv http.Server
	srv.Addr = host + ":" + port

	go func() {
		stdlog.Println(srv.ListenAndServe())
		cancel()
	}()

	go func() {
		fmt.Printf("%v started, Press any key to shut down the service", serviceName)
		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancel()
	}()
	return ctx
}
