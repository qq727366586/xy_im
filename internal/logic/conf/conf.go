package conf

import (
	"flag"
	"github.com/spf13/viper"
)

var (
	confPath string
	Conf     *Config
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

func init() {
	flag.StringVar(&confPath, "conf", "./../../config/logic.yaml", "default config path")
}

func Init() (err error) {
	Conf = &Config{}
	config := viper.New()
	config.SetConfigFile(confPath)
	config.SetConfigType("yaml")
	if err = config.ReadInConfig(); err != nil {
		return
	}
	if err = config.Unmarshal(Conf); err != nil {
		return
	}
	return
}
