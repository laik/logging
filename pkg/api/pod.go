package api

import (
	"fmt"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/service"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/watch"
)

var _ IReconcile = &Pod{}

type Pod struct {
	ns string
	service.IService
}

func NewPod(ns string, service service.IService) IReconcile {
	return &Pod{ns, service}
}

func (s *Pod) handle(slackTask *v1.SlackTask) error {
	return s.ApplySlackTask(s.ns, slackTask)
}

func existServiceName(pod *corev1.Pod) bool {
	_, exist := pod.GetLabels()["app"]
	return exist
}

func slackTaskFromPod(_type watch.EventType, pod *corev1.Pod) *v1.SlackTask {
	ips := make([]string, 0)
	for _, ip := range pod.Status.PodIPs {
		ips = append(ips, ip.IP)
	}
	return &v1.SlackTask{
		Spec: v1.SlackTaskSpec{
			Type:        _type,
			Ns:          pod.GetNamespace(),
			ServiceName: pod.GetLabels()["app"],
			Pod:         pod.GetName(),
			Ips:         ips,
			Offset:      0,
		},
	}

}

func (s *Pod) Run(errors chan error) {
	slack, err := s.GetSlack(s.ns)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			goto NEXT
		}
		errors <- err
		return
	}

NEXT:
	podList, resourceVersion, err := s.ListPod(s.ns, slack.Spec.Selector)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			goto NEXT1
		}
		errors <- err
		return
	}
NEXT1:
	for _, pod := range podList {
		if err := s.handle(slackTaskFromPod(watch.Added, pod)); err != nil {
			errors <- err
			return
		}
	}

	podWatchChan, err := s.WatchPod(s.ns, resourceVersion, slack.Spec.Selector)
	if err != nil {
		errors <- err
		return
	}
	for {
		podEvt, ok := <-podWatchChan
		if !ok {
			errors <- fmt.Errorf("watch slack task channel failed")
			return
		}
		pod := &corev1.Pod{}
		if err := core.Convert(podEvt.Object, pod); err != nil {
			errors <- err
			return
		}
		if !existServiceName(pod) {
			continue
		}

		switch podEvt.Type {
		case watch.Added, watch.Modified, watch.Deleted:
			if err := s.handle(slackTaskFromPod(podEvt.Type, pod)); err != nil {
				errors <- err
				return
			}
		default:
		}
	}
}
