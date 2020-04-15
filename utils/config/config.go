package config

import (
	"errors"
	"fmt"

	"github.com/zdnscloud/cement/configure"
	"gopkg.in/yaml.v2"

	//"flag"
	"io/ioutil"
	"log"
	"reflect"
)

type ConfigureOwner interface {
	ReloadConfig(*VanguardConf)
}

var (
	YAML_CONFIG_FILE = "/etc/vanguard/vanguard.conf"
	//configFile                    string
	ErrConfigureObjectIsNotStruct = errors.New("configure object isn't struct")
	ErrRequiredFieldIsEmpty       = errors.New("required filed hasn't been set")
)

func ReloadConfig(o interface{}, conf *VanguardConf) {
	if owner, ok := o.(ConfigureOwner); ok {
		owner.ReloadConfig(conf)
	}
}

/*
configfile example:
$ cat /etc/vanguard/vanguard.conf
server:
    kafka: 10.0.0.55:9202
    agent: 10.0.0.55:9202
    db: 10.0.0.559202
localhost:
    role: controller
    ip: 10.0.0.55
*/

type KafkaConf struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	TopicNode string `yaml:"topic_node"`
}
type PrometheusConf struct {
	IP   string `yaml:"ip"`
	Port string `yaml:"port"`
}

type LocalConf struct {
	IsController bool   `yaml:"isController"` // 3 roles: controller, dhcp, dns
	IsDHCP       bool   `yaml:"isDHCP"`       // 3 roles: controller, dhcp, dns
	IsDNS        bool   `yaml:"isDNS"`        // 3 roles: controller, dhcp, dns
	IP           string `yaml:"ip"`
	Hostname     string `yaml:"hostname"`
	ParentIP     string `yaml:"parent_ip"`
	//PromHost string `yaml:"prom_host"`
	//PromPort string `yaml:"prom_port"`
	//State  uint  `yaml:"state"`   // 1 online 0 offline
	//OnTime int64 `yaml:"on_time"` // timestamp of the nearest online time
}

type ServerConf struct {
	Kafka      KafkaConf      `yaml:"kafka"`
	Prometheus PrometheusConf `yaml:"prometheus"`
	Agent      string         `yaml:"agent"`
	DHCPGrpc   string         `yaml:"dhcpgrpc"`
	GrpcPort   string         `yaml:"grpcport"`
}

type VanguardConf struct {
	Path      string     `yaml:"-"`
	Localhost LocalConf  `yaml:"localhost"`
	Server    ServerConf `yaml:"server"`
}

func Load(config interface{}, file string) error {
	if err := processFile(config, file); err != nil {
		return err
	}
	return processTags(config)
}

func processFile(config interface{}, file string) error {
	if data, err := ioutil.ReadFile(file); err != nil {
		return err
	} else {
		return yaml.Unmarshal(data, config)
	}
}

func processTags(config interface{}) error {
	value := reflect.Indirect(reflect.ValueOf(config))
	if value.Kind() != reflect.Struct {
		return ErrConfigureObjectIsNotStruct
	}

	typ := value.Type()
	fmt.Println("type: ", typ)
	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldValue := value.Field(i)

		for fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			if err := processTags(fieldValue.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Slice:
			for i := 0; i < fieldValue.Len(); i++ {
				if reflect.Indirect(fieldValue.Index(i)).Kind() == reflect.Struct {
					if err := processTags(fieldValue.Index(i).Addr().Interface()); err != nil {
						return err
					}
				}
			}
		default:
			if isBlank := reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(fieldValue.Type()).Interface()); isBlank {
				if value := fieldType.Tag.Get("default"); value != "" {
					if err := yaml.Unmarshal([]byte(value), fieldValue.Addr().Interface()); err != nil {
						return err
					}
				} else if fieldType.Tag.Get("required") == "true" {
					return ErrRequiredFieldIsEmpty
				}
			}
		}
	}
	return nil
}
func (conf *VanguardConf) Reload() error {
	var newConf VanguardConf
	if err := configure.Load(&newConf, conf.Path); err != nil {
		return err
	}
	newConf.Path = conf.Path
	*conf = newConf

	return nil
}
func LoadConfig(path string) (*VanguardConf, error) {
	var conf VanguardConf
	conf.Path = path
	if err := conf.Reload(); err != nil {
		return nil, err
	}

	return &conf, nil
}

/*func init() {
	flag.Parse()
	flag.StringVar(&configFile, "c", YAML_CONFIG_FILE, "configure file path")
}*/

func GetConfig(confPath string) *VanguardConf {
	conf, err := LoadConfig(confPath)
	if err != nil {
		panic(PANIC_CONFIG_FILE + err.Error())
	}
	log.Println("this host ip: ", conf.Localhost.IP)
	return conf
}

func GetLocalIP(confPath string) string {

	ip := GetConfig(confPath).Localhost.IP
	log.Println("in GetLocalIP(), localhost ip: ")

	return ip
}
