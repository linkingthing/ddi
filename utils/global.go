package utils

//global vars defined here
var (
	PromServer        = "10.0.0.24" //prometheus server ip
	PromPort          = "9090"      //prometheus server port
	PromLocalhost     = "localhost" //prometheus localhost ip, server is localhost, node is ip
	PromLocalPort     = "9100"      //prometheus localhost ip
	PromLocalInstance = PromLocalhost + ":" + PromLocalPort
)
