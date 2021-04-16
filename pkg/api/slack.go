package api

import (
	"fmt"
	stack "github.com/pkg/errors"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/service"
	"k8s.io/apimachinery/pkg/watch"
)

var _ IReconcile = &Slack{}

type Slack struct {
	ns string
	service.IService
}

func NewSlack(ns string, service service.IService) IReconcile {
	return &Slack{ns, service}
}

func (s *Slack) Run(errors chan error) {
	var slack *v1.Slack
	var err error

	slack, err = s.GetSlack(s.ns)
	if err != nil {
		errors <- stack.WithStack(err)
		return
	}

	if slack == nil {
		if slack, err = s.Create(s.ns); err != nil {
			errors <- stack.WithStack(err)
			return
		}
	}

	slackChannel, err := s.WatchSlack(s.ns, slack.GetResourceVersion())
	if err != nil {
		errors <- stack.WithStack(err)
		return
	}

	for {
		slackEvt, ok := <-slackChannel
		if !ok {
			errors <- fmt.Errorf("failed to watch slack")
			return
		}
		switch slackEvt.Type {
		case watch.Deleted:
			if slack, err = s.Create(s.ns); err != nil {
				errors <- stack.WithStack(err)
				return
			}
		}
	}
}
