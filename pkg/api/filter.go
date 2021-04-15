package api

import "github.com/yametech/logging/pkg/service"

var _ IReconcile = &Slack{}

type Filter struct{ service service.IService }

func NewFilter(service service.IService) IReconcile {
	return Filter{service}
}

func (s Filter) Run(errors chan error) {

}
