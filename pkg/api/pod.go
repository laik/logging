package api

import (
	"fmt"
	stack "github.com/pkg/errors"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/service"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"time"
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "SlackTask",
			APIVersion: "logging.yamecloud.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.GetLabels()["app"],
			Namespace: pod.GetNamespace(),
		},
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
RETRY:
	slack, err := s.GetSlack(s.ns)
	if err != nil {
		if slack == nil {
			time.Sleep(3 * time.Second)
			goto RETRY
		}
		errors <- stack.WithStack(err)
		return
	}

	podList, resourceVersion, err := s.ListPod(s.ns, slack.Spec.Selector)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			goto NEXT1
		}
		errors <- stack.WithStack(err)
		return
	}

NEXT1:
	for _, pod := range podList {
		if err := s.handle(slackTaskFromPod(watch.Added, pod)); err != nil {
			errors <- stack.WithStack(err)
			return
		}
	}

	podWatchChan, err := s.WatchPod(s.ns, resourceVersion, slack.Spec.Selector)
	if err != nil {
		errors <- stack.WithStack(err)
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
			errors <- stack.WithStack(err)
			return
		}
		if !existServiceName(pod) {
			continue
		}

		switch podEvt.Type {
		case watch.Added, watch.Modified, watch.Deleted:
			if err := s.handle(slackTaskFromPod(podEvt.Type, pod)); err != nil {
				errors <- stack.WithStack(err)
				return
			}
		default:
		}
	}
}
