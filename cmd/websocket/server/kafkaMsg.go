package server

import (
	"github.com/linkingthing/ddi/utils"
	"log"
	"strconv"
	"time"
)

func getKafkaMsg() {
	log.Println("into getKafkaMsg")
	for {

		log.Println("in loop" + strconv.FormatInt(time.Now().Unix(), 10))

		utils.ConsumerProm()

		time.Sleep(checkDuration)
		//time.Sleep(20 * time.Second)
	}
}
