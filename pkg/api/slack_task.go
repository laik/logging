package api

import (
	"fmt"
	stack "github.com/pkg/errors"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/command"
	"github.com/yametech/logging/pkg/common"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/service"
	"github.com/yametech/logging/pkg/utils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
	var filter *v1.Filter
	var err error

	options := []command.Option{
		command.WithNs(s.ns),
		command.WithIPs(slackTask.Spec.Ips...),
		command.WithNodeName(slackTask.Spec.Node),
		command.WithPodName(slackTask.Spec.Pod),
		command.WithOffset(slackTask.Spec.Offset),
		command.WithServiceName(slackTask.Spec.ServiceName),
	}

	if slackTask.Spec.FilterName != "" {
		filter, err = s.GetFilter(s.ns, slackTask.Spec.FilterName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				goto NEXT
			}
			return stack.WithStack(err)
		}
	}
NEXT:

	if filter != nil {
		options = append(options, command.WithFilter(filter.Spec.MaxLength, filter.Spec.Expr))
	}

	sink, err := s.GetSink(s.ns)
	if err != nil {
		return stack.WithStack(err)
	}

	if sink == nil || sink.Spec.Address == nil || sink.Spec.Type == nil {
		fmt.Printf("%s sink not found or sink not define %v\n", common.WARN, sink)
		return nil
	}

	exist, err := utils.CheckTopicExist(*sink.Spec.Address, slackTask.Spec.ServiceName)
	if err != nil {
		return err
	}

	if !exist {
		if err := utils.CreateTopic(*sink.Spec.Address, slackTask.Spec.ServiceName, *sink.Spec.Partition); err != nil {
			return err
		}
	}

	options = append(options, command.WithOutput(fmt.Sprintf("%s:%s@%s", *sink.Spec.Type, slackTask.Spec.ServiceName, *sink.Spec.Address)))

	switch slackTask.Spec.Type {
	case watch.Added, watch.Modified:
		options = append(options, command.WithOp(command.RUN))
	case watch.Deleted:
		options = append(options, command.WithOp(command.STOP))
	}

	cmdStr, err := command.CMD(options...)
	if err != nil {
		return stack.WithStack(err)
	}

	s.broadcast.Publish(cmdStr)

	return nil
}

func (s SlackTask) Run(errors chan error) {
	slackTaskList, resourceVersion, err := s.ListSlackTask(s.ns)
	if err != nil {
		errors <- stack.WithStack(err)
		return
	}

	for _, slackTask := range slackTaskList {
		if err := s.handle(slackTask); err != nil {
			errors <- stack.WithStack(err)
			return
		}
	}

	slackTaskChan, err := s.WatchSlackTask(s.ns, resourceVersion)
	if err != nil {
		errors <- stack.WithStack(err)
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
			errors <- stack.WithStack(err)
			return
		}

		if err := s.handle(slackTask); err != nil {
			errors <- stack.WithStack(err)
			return
		}
	}

}
