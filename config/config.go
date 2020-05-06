package config

import (
	"github.com/zdnscloud/cement/configure"
)

type DDIControllerConfig struct {
	Path   string     `yaml:"-"`
	DB     DBConf     `yaml:"db"`
	Local  LocalConf  `yaml:"local"`
	Server ServerConf `yaml:"server"`
}

type DBConf struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
}

type LocalConf struct {
	IsController bool   `yaml:"is_controller"`
	IsDHCP       bool   `yaml:"is_dhcp"`
	IsDNS        bool   `yaml:"is_dns"`
	IP           string `yaml:"ip"`
	Hostname     string `yaml:"hostname"`
	ParentIP     string `yaml:"parent_ip"`
	Role         string `yaml:"role"`
}

type ServerConf struct {
	Kafka          KafkaConf      `yaml:"kafka"`
	Prometheus     PrometheusConf `yaml:"prometheus"`
	Agent          string         `yaml:"agent"`
	DHCPGrpc       string         `yaml:"dhcp_grpc"`
	GrpcPort       string         `yaml:"grpc_port"`
	GrpcDhcpv4Port string         `yaml:"grpc_dhcpv4_port"`
	GrpcDhcpv6Port string         `yaml:"grpc_dhcpv6_Port"`
}

type KafkaConf struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	TopicNode string `yaml:"topic_node"`
}

type PrometheusConf struct {
	IP   string `yaml:"ip"`
	Port string `yaml:"port"`
}

func LoadConfig(path string) (*DDIControllerConfig, error) {
	var conf DDIControllerConfig
	conf.Path = path
	if err := conf.Reload(); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (c *DDIControllerConfig) Reload() error {
	var newConf DDIControllerConfig
	if err := configure.Load(&newConf, c.Path); err != nil {
		return err
	}

	newConf.Path = c.Path
	*c = newConf
	return nil
}
