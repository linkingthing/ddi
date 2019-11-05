package dns

import (
	"github.com/linkingthing/ddi/pb"
)

func TestStartDNS(t *testing.T) {
	var config string = "options {\n\tdirectory \"/root/bindtest/\";\n\tpid-file \"named.pid\";\n\tallow-new-zones yes;\n\tallow-query {any;};\n};\n" +
		"view \"default\" {\n\tmatch-clients {\n\tany;\n\t};\n};\n" +
		"key \"rndc-key\" {\n\talgorithm hmac-sha256;\n\tsecret \"4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=\";\n};\n" +
		"controls {\n\tinet 127.0.0.1 port 953\n\tallow { 127.0.0.1; } keys { \"rndc-key\"; };\n};\n" +
		"include \"/root/bindtest/named.rfc1912.zones\";\n"
	dnsStartReq := pb.DNSStartReq{Config: config}
	StartDNS(dnsStartReq)

}
