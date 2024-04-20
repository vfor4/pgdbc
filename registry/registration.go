package registry

type Registration struct {
	ServiceName      ServiceName
	ServiceUrl       string
	RequiredServices []ServiceName
	ServiceUpdateURL string
}

type ServiceName string

const (
	LogService     = ServiceName("LogService")
	GradingService = ServiceName("GradingService")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}
