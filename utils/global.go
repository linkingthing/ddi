package utils

//global vars defined here
var (
	PromServer        = "10.0.0.24" //prometheus server ip
	PromPort          = "9090"      //prometheus server port
	PromLocalhost     = "localhost" //prometheus localhost ip, server is localhost, node is ip
	PromLocalPort     = "9100"      //prometheus localhost ip
	PromLocalInstance = PromLocalhost + ":" + PromLocalPort
	WebSocket_Port    = 13333 //ping pong check port

	EsServer = "10.0.0.69" //elasticsearch server
	EsPort   = "9200"      //elasticsearch port
	EsIndex  = "dns_log"   //elasticsearch index name

	DBAddr = "postgresql://maxroach@localhost:26257/ddi?ssl=true&sslmode=require&sslrootcert=/root/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/cockroach-v19.2.0/certs/client.maxroach.crt"

	KeaServer       = "10.0.0.31" //host ip on which kea is running
	KeaPort         = 8000        //host port on which kea is running
	NodeRole        = "controller"
	KafkaServerProm = ""
	DHCPGrpcServer  = ""
	Dhcpv4AgentAddr = "localhost:8898"
	Dhcpv6AgentAddr = "localhost:8899"
	GrpcServer      = "127.0.0.1:8888"
	DhcpHost        = "10.0.0.31"
)
