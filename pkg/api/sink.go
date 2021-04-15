package api

import "github.com/yametech/logging/pkg/service"

var _ IReconcile = &Sink{}

type Sink struct{ service service.IService }

func NewSink(service service.IService) IReconcile {
	return &Sink{service}
}

func (s *Sink) Run(errors chan error) {
	// AutoCreateSink
}
