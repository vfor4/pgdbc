package runner

import (
	"context"
	"main/services/registry"
)

func Start(ctx context.Context, host, port string, r registry.Registration, registerHandlersFunc func) {
	registerHandlersFunc()
}

func startService(ctx concontext.Context, serviceName registry.ServiceName) {
}
