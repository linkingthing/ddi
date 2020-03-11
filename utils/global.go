package utils

//global vars defined here
var (
	PromServer        = "10.0.0.24" //prometheus server ip
	PromPort          = "9090"      //prometheus server port
	PromLocalhost     = "localhost" //prometheus localhost ip, server is localhost, node is ip
	PromLocalPort     = "9100"      //prometheus localhost ip
	PromLocalInstance = PromLocalhost + ":" + PromLocalPort
	WebSocket_Port    = 3333 //ping pong check port

	EsServer = "10.0.0.69" //elasticsearch server
	EsPort   = "9200"      //elasticsearch port
	EsIndex  = "dns_log"   //elasticsearch index name

	KeaServer = "10.0.0.31" //host ip on which kea is running
	KeaPort   = 8000        //host port on which kea is running

)
