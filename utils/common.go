package utils

import (
	"github.com/linkingthing/ddi/utils/config"
	"log"
)

func GetLocalIP() string {

	ip := config.GetConfig().Localhost.IP
	log.Println("in GetLocalIP(), localhost ip: ")

	return ip
}
