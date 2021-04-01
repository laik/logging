package command

import "k8s.io/apimachinery/pkg/util/json"

type Op = string

const (
	RUN   Op = "run"
	STOP  Op = "stop"
	HELLO Op = "hello"
)

type Pod struct {
	Name     string   `json:"name"`
	NodeName string   `json:"node_name"`
	Ips      []string `json:"ips"`
	Offset   uint64   `json:"offset"`
}

func NewPod() *Pod {
	return &Pod{Ips: make([]string, 0)}
}

func (p *Pod) SetName(name string) *Pod {
	p.Name = name
	return p
}

func (p *Pod) SetNodeName(name string) *Pod {
	p.NodeName = name
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
	Op     Op     `json:"op"`
	Ns     string `json:"ns"`
	Rules  string `json:"rules"`
	Output string `json:"output"`
	Pods   []Pod  `json:"pods"`
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

func (c *Cmd) Run() *Cmd {
	c.Op = RUN
	return c
}

func (c *Cmd) Stop() *Cmd {
	c.Op = STOP
	return c
}

func (c *Cmd) Hello() *Cmd {
	c.Op = HELLO
	return c
}

func (c *Cmd) SetOutput(o string) *Cmd {
	c.Output = o
	return c
}

func (c *Cmd) SetRule(o string) *Cmd {
	c.Rules = o
	return c
}

func (c *Cmd) SetNs(o string) *Cmd {
	c.Ns = o
	return c
}

func (c *Cmd) AddPod(pod *Pod) *Cmd {
	position := -1
	for index, v := range c.Pods {
		if v.Name == pod.Name {
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
		if v.Name != pod.Name {
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
