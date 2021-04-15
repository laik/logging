package api

import "github.com/yametech/logging/pkg/service"

var _ IReconcile = &Slack{}

type Slack struct{ service.IService }

func NewSlack(service service.IService) IReconcile {
	return &Slack{service}
}

func (s *Slack) Run(errors chan error) {}
