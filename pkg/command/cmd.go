package command

import (
	"k8s.io/apimachinery/pkg/util/json"
)

type Op = string

const (
	RUN   Op = "run"
	STOP  Op = "stop"
	HELLO Op = "hello"
)

type Option func(*cmd)

type cmd struct {
	Op Op     `json:"op"`
	Ns string `json:"ns"`

	Output      string   `json:"output"`
	ServiceName string   `json:"service_name"`
	PodName     string   `json:"pod_name"`
	NodeName    string   `json:"node_name"`
	Ips         []string `json:"ips"`
	Offset      uint64   `json:"offset"`

	Filter map[string]interface{} `json:"filter"`
}

func CMD(ops ...Option) (string, error) {
	cmd := &cmd{}
	for _, op := range ops {
		op(cmd)
	}
	if cmd.Filter == nil {
		cmd.Filter = make(map[string]interface{})
		cmd.Filter["max_length"] = 1e9
		cmd.Filter["expr"] = "*"
	}
	return cmd.ToString()
}

func (c *cmd) ToString() (string, error) {
	bs, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func WithOp(op Op) Option {
	return func(cmd *cmd) {
		cmd.Op = op
	}
}

func WithOffset(offset uint64) Option {
	return func(cmd *cmd) {
		cmd.Offset = offset
	}
}

func WithNs(ns string) Option {
	return func(cmd *cmd) {
		cmd.Ns = ns
	}
}

func WithFilter(maxLength uint64, expr string) Option {
	return func(cmd *cmd) {
		if cmd.Filter == nil {
			cmd.Filter = make(map[string]interface{})
		}
		cmd.Filter["max_length"] = maxLength
		cmd.Filter["expr"] = expr
	}
}

func WithOutput(o string) Option {
	return func(cmd *cmd) {
		cmd.Output = o
	}
}

func WithServiceName(s string) Option {
	return func(cmd *cmd) {
		cmd.ServiceName = s
	}
}

func WithNodeName(s string) Option {
	return func(cmd *cmd) {
		cmd.NodeName = s
	}
}

func WithIPs(ip ...string) Option {
	return func(cmd *cmd) {
		if cmd.Ips == nil {
			cmd.Ips = make([]string, 0)
		}
		cmd.Ips = append(cmd.Ips, ip...)
	}
}

func WithPodName(pod string) Option {
	return func(cmd *cmd) {
		cmd.PodName = pod
	}
}
