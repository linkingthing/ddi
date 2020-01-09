//check node register info every day
package main

import (
	"github.com/linkingthing/ddi/utils"
	"log"
	"strconv"
	"time"
)

var checkDuration = 24 * time.Hour

func main() {

	for true {
		log.Println("in loop" + strconv.FormatInt(time.Now().Unix(), 10))

		utils.ConsumerProm()

		time.Sleep(checkDuration)
		//time.Sleep(20 * time.Second)
	}
}
