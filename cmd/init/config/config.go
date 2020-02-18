package config

import (
	"errors"
	"flag"
	"github.com/zdnscloud/cement/configure"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"reflect"
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
	configFile                    string
	ErrConfigureObjectIsNotStruct = errors.New("configure object isn't struct")
	ErrRequiredFieldIsEmpty       = errors.New("required filed hasn't been set")
)

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
