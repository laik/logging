package api

import (
	"fmt"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/command"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/service"
	"k8s.io/apimachinery/pkg/watch"
)

var _ IReconcile = &SlackTask{}

type SlackTask struct {
	ns string

	broadcast *Broadcast
	service.IService
}

func NewSlackTask(ns string, broadcast *Broadcast, service service.IService) IReconcile {
	return &SlackTask{ns, broadcast, service}
}

func (s *SlackTask) handle(slackTask *v1.SlackTask) error {
	filter, err := s.GetFilter(s.ns, slackTask.Spec.FilterName)
	if err != nil {
		return err
	}
	options := []command.Option{
		command.WithNs(s.ns),
		command.WithIPs(slackTask.Spec.Ips...),
		command.WithNodeName(slackTask.Spec.Node),
		command.WithPodName(slackTask.Spec.Pod),
		command.WithOffset(slackTask.Spec.Offset),
		command.WithFilter(filter.Spec.MaxLength, filter.Spec.Expr),
	}

	switch slackTask.Spec.Type {
	case watch.Added, watch.Modified:
		options = append(options, command.WithOp(command.RUN))
	case watch.Deleted:
		options = append(options, command.WithOp(command.STOP))
	}

	cmdStr, err := command.CMD(options...)
	if err != nil {
		return err
	}
	s.broadcast.Publish(cmdStr)
	return nil
}

func (s SlackTask) Run(errors chan error) {
	slackTaskList, resourceVersion, err := s.ListSlackTask(s.ns)
	if err != nil {
		errors <- err
		return
	}

	for _, slackTask := range slackTaskList {
		if err := s.handle(slackTask); err != nil {
			errors <- err
			return
		}
	}

	slackTaskChan, err := s.WatchSlackTask(s.ns, resourceVersion)
	if err != nil {
		errors <- err
		return
	}

	for {
		slackTaskEvt, ok := <-slackTaskChan
		if !ok {
			errors <- fmt.Errorf("watch slack task channel failed")
			return
		}
		slackTask := &v1.SlackTask{}
		if err := core.Convert(slackTaskEvt.Object, slackTask); err != nil {
			errors <- err
			return
		}

		if err := s.handle(slackTask); err != nil {
			errors <- err
			return
		}
	}

}
