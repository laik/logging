package client

type Params map[string]interface{}

type HttpClient interface {
	Set(Params) HttpClient
	Post(uri string) error
}
