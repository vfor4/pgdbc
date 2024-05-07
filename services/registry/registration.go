package registry

type Registration struct {
	ServiceName ServiceName
}

type ServiceName string

const (
	LogService = ServiceName("LogService")
)
