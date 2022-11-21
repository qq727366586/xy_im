package conf

var (
	Conf *Config
)

type Config struct {
	Env       *Env
	RPCServer *RPCServer
}

type Env struct {
	DeployEnv string
	Host      string
}
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           int
	IdleTimeout       int
	MaxLifeTime       int
	ForceCloseWait    int
	KeepAliveInterval int
	KeepAliveTimeout  int
}
