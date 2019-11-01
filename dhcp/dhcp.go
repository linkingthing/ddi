package dhcp

import (
	"fmt"
	"os/exec"
)

func Output() {

	fmt.Print("%d, test", 10)
}

func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}

//service: dhcp4, dhcp6, ctrl_agent, ddns
func StartDHCP(service string) error {
	var command string = "ps -eaf|grep kea-" + service + "|grep -v grep"
	if ret, err := cmd(command); err != nil {
		fmt.Println(service + " hasn't started!start now!")

		if ret, err := cmd(service); err != nil {
			fmt.Printf("start fail! return message:%s\n", ret)
			fmt.Println(err)
			return err
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

//service: dhcp4, dhcp6, ctrl_agent, ddns
func StopDHCP(service string) error {
	var command string = "ps -eaf|grep kea-" + service + "|grep -v grep"
	if ret, err := cmd(command); err != nil {
		fmt.Println(service + " hasn't started!start now!")

		if ret, err := cmd(service); err != nil {
			fmt.Printf("start fail! return message:%s\n", ret)
			fmt.Println(err)
			return err
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

func CreateSubnet(v46 string) error {

	return nil
}
func UpdateSubnet(v46 string) error {

	return nil
}
