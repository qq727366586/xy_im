package conf

import (
	"flag"
	"github.com/spf13/viper"
	"os"
)

var (
	confPath string
	host     string
	Conf     *Config
)

type Config struct {
	Env       *Env
	RPCClient *RPCClient
	TCP       *TCP
	Websocket *Websocket
	Protocol  *Protocol
	Bucket    *Bucket
}

type Env struct {
	DeployEnv string
	Host      string
}

type RPCClient struct {
	Dial    int
	Timeout int
	Bind    string
}

type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	KeepAlive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

type Websocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

type Protocol struct {
	Timer            int
	TimerSize        int
	SvrProto         int
	CliProto         int
	HandshakeTimeout int
}

type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

func init() {
	defHost, _ := os.Hostname()
	flag.StringVar(&confPath, "conf", "./../../config/comet.yaml", "default config path")
	flag.StringVar(&host, "host", defHost, "machine hostname, also can use default machine hostname")

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
