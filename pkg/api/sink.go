package api

import (
	"fmt"
	stack "github.com/pkg/errors"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/service"
	"k8s.io/apimachinery/pkg/watch"
)

var _ IReconcile = &Sink{}

type Sink struct {
	ns string
	service.IService
}

func NewSink(ns string, service service.IService) IReconcile {
	return &Sink{ns, service}
}

func (s *Sink) Run(errors chan error) {
	var sink *v1.Sink
	var err error

	sink, err = s.GetSink(s.ns)
	if err != nil {
		errors <- stack.WithStack(err)
		return
	}

	if sink == nil {
		if sink, err = s.CreateSink(s.ns); err != nil {
			errors <- stack.WithStack(err)
			return
		}
	}

	slackChannel, err := s.WatchSlack(s.ns, sink.GetResourceVersion())
	if err != nil {
		errors <- stack.WithStack(err)
		return
	}

	for {
		slackEvt, ok := <-slackChannel
		if !ok {
			errors <- fmt.Errorf("failed to watch sink")
			return
		}
		switch slackEvt.Type {
		case watch.Deleted:
			if sink, err = s.CreateSink(s.ns); err != nil {
				errors <- stack.WithStack(err)
				return
			}
		}
	}
}
