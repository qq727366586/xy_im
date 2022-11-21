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
	TCP       *TCP
	Websocket *Websocket
	Protocol  *Protocol
	Bucket    *Bucket
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
	flag.StringVar(&confPath, "conf", "./../../config/comet.yaml", "default config path")
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
