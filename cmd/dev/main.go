package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

const cmd = `{"op":"run","ns":"default","service_name":"xx_service","filter":{"max_length":1024,"expr":"[INFO]"},"output":"fake_output","node_name":"node1","pod_name" :"pod-12345","container":"nginx1","ips":[ "127.0.0.1"],"offset":0}`

func StreamData(c *gin.Context) {
	chanStream := make(chan string, 10)
	go func() {
		for {
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

	route.GET("/:node", StreamData)

	if err := route.Run("0.0.0.0:9999"); err != nil {
		panic(err)
	}
}

func main() {
	example()
}
