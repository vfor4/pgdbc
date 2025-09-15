package elephas

type ReplicationOpts string

const (
	RepTrue  ReplicationOpts = "true"
	RepFalse ReplicationOpts = "false"
	RepDB    ReplicationOpts = "database"
)

type StartupParamName string

const (
	User     StartupParamName = "user"
	Database StartupParamName = "database"
	// Opts StartupParamName = "Options"
	Replication StartupParamName = "replication"
)

type Config struct {
	// Network type, either tcp or unix.
	// Default is tcp
	Network string
	// TCP host:port or Unix socket depending on Network.
	Addr        string
	User        string
	Password    string
	Database    string
	Replication string
}
