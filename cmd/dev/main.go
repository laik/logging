package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

const cmd = `{"op":"run","ns":"kube-system","service_name":"echoer-api","filter":{"max_length":"1024","expr":"[INFO]"},"output":"fake_output","pods":[{"node":"node1","pod":"echoer-api-86c648d678-z2p9p","container":"echoer-api-86c648d678-z2p9p","ips":["127.0.0.1"],"offset":0}]}`

func StreamData(c *gin.Context) {
	chanStream := make(chan string, 10)
	go func() {
		for {
			//chanStream <- `{"op":"run","ns":"kube-system","service_name":"kube-controller-manager","rules":"","output":"kafka:test2@10.200.100.200:9092","pods":[{"node":"node1","pod":"kube-controller-manager-node1","ips":["10.200.64.10"],"offset":0}]}`
			//chanStream <- `{"op":"run","ns":"kube-system","service_name":"harvest-pf9wg","rules":"","output":"kafka:test2@10.200.100.200:9092","pods":[{"node":"node1","pod":"harvest-pf9wg","ips":["127.0.0.1"],"offset":0}]}`
			//chanStream <- `{"op":"run","ns":"kube-system","service_name":"compass","rules":"","output":"kafka:test2@10.200.100.200:9092","pods":[{"node":"node1","pod":"compass-64777666c6-95hbk","ips":["127.0.0.1"],"offset":0}]}`
			//chanStream <- `{"op":"run","ns":"finance-dev","service_name":"sky-fcms-web-ui","rules":"","output":"fake_output","pods":[{"node":"node1","pod":"sky-fcms-web-ui-0-b-0","ips":["127.0.0.1"],"offset":0}]}`
			//chanStream <- `{"op":"run","ns":"kube-system","service_name":"echoer-api","rules":"","output":"kafka:test3@10.200.100.200:9092","pods":[{"node":"node1","pod":"echoer-api-86c648d678-z2p9p","ips":["127.0.0.1"],"offset":0}]}`
			//chanStream <- `{"op":"run","ns":"kube-system","service_name":"kube-apiserver","rules":"","output":"kafka:test2@10.200.100.200:9092","pods":[{"node":"node1","pod":"kube-apiserver-node1","ips":["127.0.0.1"],"offset":0}]}`
			//chanStream <- `{"op":"run","ns":"kube-system","service_name":"workload","rules":"","output":"fake_output","pods":[{"node":"cg-compass-dev-01","pod": "workload-56f8bdb47d-78sx8","ips":["127.0.0.1"],"offset":0}]}`
			//chanStream <- `{"op":"run","ns":"kube-system","service_name":"ovn-central","rules":"","output":"kafka:test2@10.200.100.200:9092","pods":[{"node":"node1","pod": "ovn-central-9f5754ccc-7nz77","ips":["127.0.0.1"],"offset":0}]}`
			chanStream <- cmd
			time.Sleep(time.Second * 15)
		}
	}()
	c.Stream(func(w io.Writer) bool {
		c.SSEvent("", <-chanStream)
		return true
	})
}

func example() {
	route := gin.Default()

	route.GET("/", StreamData)

	if err := route.Run("0.0.0.0:9999"); err != nil {
		panic(err)
	}
}

func main() {
	example()
}
