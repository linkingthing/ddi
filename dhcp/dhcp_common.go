package dhcp

import (
	"github.com/linkingthing/ddi/pb"
	"os/exec"
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
