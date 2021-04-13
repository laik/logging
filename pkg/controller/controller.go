package controller

import (
	"fmt"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/common"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/service"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type PodMonitorController struct {
	ns              string
	service         service.IService
	resourceVersion string
}

func NewPodMonitorController(ns string, service service.IService) *PodMonitorController {
	return &PodMonitorController{
		ns, service, "0",
	}
}

func (p *PodMonitorController) podToTask(pod *corev1.Pod) v1.Task {
	return v1.Task{
		Ns:          pod.Namespace,
		ServiceName: pod.Labels["app"],
		Filter:      v1.Filter{},
		Pods:        []v1.Pod{},
	}
}

func (p *PodMonitorController) handle(podEvt watch.Event) error {
	pod := corev1.Pod{}
	if err := core.Convert(podEvt.Object, &pod); err != nil {
		return err
	}

	switch podEvt.Type {
	case watch.Added, watch.Modified:
		if pod.Status.Phase == corev1.PodRunning {
			slack, err := p.service.GetSlack(p.ns, fmt.Sprintf(common.NamespaceSlackName, p.ns))
			if err != nil {
				return err
			}
			slack.Spec.AddTasks = append(slack.Spec.AddTasks, p.podToTask(&pod))
			if err := p.service.UpdateSlack(p.ns, slack.Name, slack); err != nil {
				return err
			}
		}
	case watch.Deleted:
		slack, err := p.service.GetSlack(p.ns, fmt.Sprintf(common.NamespaceSlackName, p.ns))
		if err != nil {
			return err
		}
		slack.Spec.DeleteTasks = append(slack.Spec.DeleteTasks, p.podToTask(&pod))
		if err := p.service.UpdateSlack(p.ns, slack.Name, slack); err != nil {
			return err
		}
	case watch.Error:
		return fmt.Errorf("PodMonitorController handle watch error")
	}

	return nil
}

func (p *PodMonitorController) Start() <-chan error {
	errors := make(chan error)
	watchPodChan, err := p.service.WatchPod(p.ns, p.resourceVersion, "")
	if err != nil {
		errors <- err
	}
	go func() {
		for {
			select {
			case evt, ok := <-watchPodChan:
				if !ok {
					errors <- fmt.Errorf("%s", "watch pod task chan error")
				}
				if err := p.handle(evt); err != nil {
					errors <- err
				}
			}
		}
	}()

	return errors
}
