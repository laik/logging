package command

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/json"
)

type Op = string

const (
	RUN   Op = "run"
	STOP  Op = "stop"
	HELLO Op = "hello"
)

type Pod struct {
	PodName   string   `json:"pod_name"`
	NodeName  string   `json:"node_name"`
	Container string   `json:"container"`
	Ips       []string `json:"ips"`
	Offset    uint64   `json:"offset"`
}

func NewPod() *Pod {
	return &Pod{Ips: make([]string, 0)}
}

func (p *Pod) SetName(name string) *Pod {
	p.PodName = name
	return p
}

func (p *Pod) SetNodeName(name string) *Pod {
	p.NodeName = name
	return p
}

func (p *Pod) SetIPs(ips ...string) *Pod {
	for _, ip := range ips {
		p.AddIp(ip)
	}
	return p
}

func (p *Pod) AddIp(ip string) *Pod {
	if stringSliceContains(p.Ips, ip) {
		return p
	}
	p.Ips = append(p.Ips, ip)
	return p
}

func (p *Pod) SetOffset(offset uint64) *Pod {
	p.Offset = offset
	return p
}

type Cmd struct {
	Op Op     `json:"op"`
	Ns string `json:"ns"`

	Filter      map[string]string `json:"filter"`
	Output      string            `json:"output"`
	ServiceName string            `json:"service_name"`

	Pods []Pod `json:"pods"`
}

func NewCmd() *Cmd {
	return &Cmd{
		Pods: make([]Pod, 0),
	}
}

func (c *Cmd) ToString() (string, error) {
	bs, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (c *Cmd) SetOp(op Op) *Cmd {
	c.Op = op
	return c
}

func (c *Cmd) Run() *Cmd {
	return c.SetOp(RUN)
}

func (c *Cmd) Stop() *Cmd {
	return c.SetOp(STOP)
}

func (c *Cmd) Hello() *Cmd {
	return c.SetOp(HELLO)
}

func (c *Cmd) SetOutput(o string) *Cmd {
	c.Output = o
	return c
}

func (c *Cmd) SetFilter(maxLength uint64, expr string) *Cmd {
	c.Filter["max_length"] = fmt.Sprintf("%d", maxLength)
	c.Filter["expr"] = expr
	return c
}

func (c *Cmd) SetNs(o string) *Cmd {
	c.Ns = o
	return c
}

func (c *Cmd) SetServiceName(o string) *Cmd {
	c.ServiceName = o
	return c
}

func (c *Cmd) AddPod(pod *Pod) *Cmd {
	position := -1
	for index, v := range c.Pods {
		if v.PodName == pod.PodName {
			position = index
		}
	}
	if position == -1 {
		c.Pods = append(c.Pods, *pod)
		return c
	}
	c.Pods[position] = *pod

	return c
}

func (c *Cmd) AddIp(pod *Pod, ip string) *Cmd {
	for _, v := range c.Pods {
		if v.PodName != pod.PodName {
			continue
		}
		v.AddIp(ip)
	}
	return c
}

func stringSliceContains(slice []string, item string) bool {
	for _, v := range slice {
		if item == v {
			return true
		}
	}
	return false
}
