package config

import (
	"flag"
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

func LoadConfig(path string) (*VanguardConf, error) {
	var conf VanguardConf
	conf.Path = path
	if err := conf.Reload(); err != nil {
		return nil, err
	}

	return &conf, nil
}
func init() {
	flag.Parse()
	flag.StringVar(&configFile, "c", "/etc/vanguard/vanguard.conf", "configure file path")
}

func main() {
	conf, err := LoadConfig(configFile)
	if err != nil {
		panic("load configure file failed:" + err.Error())
	}
	log.Println("conf: ", conf)
}
