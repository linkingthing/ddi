package dns

import (
	"fmt"
	"github.com/ben-han-cn/cement/shell"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

/*func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}*/

func WriteWithIOutil(name, content string) error {
	if err := ioutil.WriteFile(name, []byte(content), 0644); err != nil {
		return fmt.Errorf("write data into %s fail, %w", name, err)
	}
	return nil
}

func FileRead(name string, size int) (fileContent string, err error) {
	f, err := os.Open(name)
	if err == nil {
		return fileContent, err
	}
	defer f.Close()

	buf := make([]byte, size)
	if n, err := f.Read(buf); err == nil {
		fmt.Println("The number of bytes read:"+strconv.Itoa(n), "Buf length:"+strconv.Itoa(len(buf)))
		fileContent = string(buf)
	}
	return fileContent, nil
}

func PutDBKeyValue(inputData map[string]string) error {
	return nil
}

func StartDNS(configString string) error {
	name := "/etc/named.conf"
	if len(configString) > 0 {
		WriteWithIOutil(name, configString)

		db, err := bolt.Open("my.db", 0600, nil)
		if err != nil {
			return err
		}
		defer db.Close()
		err = db.Update(func(tx *bolt.Tx) error {

			b := tx.Bucket([]byte("binddb"))
			if b == nil {

				b, err = tx.CreateBucket([]byte("binddb"))
				if err != nil {

					log.Fatal(err)
				}

			}
			err = b.Put([]byte("named.conf"), []byte(configString))
			if err != nil {
				return err
			}
			fmt.Println("Put the key value success into binddb!")

			return nil
		})
	}
	var command string = "ps -eaf|grep named|grep -v grep"
	var ret string
	var err error
	if ret, err = shell.Shell(command); err != nil {
		fmt.Println("hasn't started!start now!")
		var start_cmd string = "named"

		if ret, err = shell.Shell(start_cmd); err != nil {
			fmt.Printf("start fail! return message:%s\n", ret)
			fmt.Println(err)
		} else {
			fmt.Printf("start success! return message:%s\n", ret)
		}

	} else if len(ret) > 0 {
		fmt.Println("had started!")
	} else {
		fmt.Println("Nothing done!")
	}
	return nil
}
