package config

import (
	"flag"
	"github.com/zdnscloud/vanguard/config"
	"log"
)

type ConfigureOwner interface {
	ReloadConfig(*VanguardConf)
}

func ReloadConfig(o interface{}, conf *VanguardConf) {
	if owner, ok := o.(ConfigureOwner); ok {
		owner.ReloadConfig(conf)
	}
}

type VanguardConf struct {
	Path string `yaml:"-"`
	Role string `yaml:"role"`
	IP   string `yaml:"ip"`
}

var (
	configFile string
)

func init() {
	flag.Parse()
	flag.StringVar(&configFile, "c", "/etc/vanguard/vanguard.conf", "configure file path")
}

func main() {
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		panic("load configure file failed:" + err.Error())
	}
	log.Println("conf: ", conf)
}
