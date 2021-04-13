package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

type Params map[string]interface{}

type HttpClient interface {
	IP(string) HttpClient
	Port(string) HttpClient
	Get(path string) (string, error)
}

func NewHttpClient() HttpClient {
	return &httpClient{}
}

var _ HttpClient = &httpClient{}

type httpClient struct {
	ip   string
	port string
}

func (h httpClient) IP(ip string) HttpClient {
	h.ip = ip
	return h
}

func (h httpClient) Port(port string) HttpClient {
	h.port = port
	return h
}

func (h httpClient) Get(path string) (string, error) {
	url := fmt.Sprintf("http://%s:%s%s", h.ip, h.port, path)
	client := resty.New()
	resp, err := client.R().SetHeader("Accept", "application/json").Get(url) //default json
	if err != nil {
		return "", fmt.Errorf("get url %s error %s", url, err)
	}
	return resp.String(), nil
}
