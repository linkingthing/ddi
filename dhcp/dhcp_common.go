package dhcp

import (
	"log"
	"os/exec"

	"github.com/linkingthing/ddi/pb"
)

func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}

func ConvertOptionsFromPb(options []*pb.Option) ([]*Option, error) {

	var ops = []*Option{}
	if len(options) > 0 {
		for _, op := range options {
			var o *Option
			o.AlwaysSend = op.AlwaysSend
			o.Code = op.Code
			o.CsvFormat = op.CsvFormat
			o.Data = op.Data
			o.Name = op.Name
			o.Space = op.Space
			ops = append(ops, o)
		}
	}

	return ops, nil
}
func CreateOptionsFromPb(gateway string, dnsServer string) ([]*Option, error) {

	var ops = []*Option{}
	log.Println("in CreateOptionsFromPb req.gateway: ", gateway)
	if len(gateway) > 0 {
		option := Option{
			Name: "routers",
			Data: gateway,
		}
		ops = append(ops, &option)
	}
	if len(dnsServer) > 0 {
		option := Option{
			Name: "domain-name-servers",
			Data: dnsServer,
		}
		ops = append(ops, &option)
	}
	log.Println("in CreateOptionsFromPb ops: ", ops)
	return ops, nil
}
func CreatePbOptions(options []*Option) []*pb.Option {
	var pbOptions []*pb.Option
	for _, op := range options {
		var pbOption pb.Option
		pbOption.Name = op.Name
		pbOption.Data = op.Data
		pbOptions = append(pbOptions, &pbOption)
	}
	return pbOptions
}
