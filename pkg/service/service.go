package service

import (
	"fmt"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/common"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/datasource"
	"github.com/yametech/logging/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type IService interface {
	WatchSlack(ns, resourceVersion string) (<-chan watch.Event, error)
	GetSlack(ns string) (*v1.Slack, error)
	UpdateSlack(slack *v1.Slack) error

	ListPod(ns string, selector string) ([]*corev1.Pod, string, error)
	WatchPod(ns string, resourceVersion, selector string) (<-chan watch.Event, error)

	GetSink(ns string) (*v1.Sink, error)
	GetFilter(ns, name string) (*v1.Filter, error)

	ListSlackTask(ns string) ([]*v1.SlackTask, string, error)
	WatchSlackTask(ns, resourceVersion string) (<-chan watch.Event, error)
	ApplySlackTask(ns string, slackTask *v1.SlackTask) error
}

type Service struct {
	datasource datasource.IDataSource
}

func (s *Service) ApplySlackTask(ns string, slackTask *v1.SlackTask) error {
	unstructuredData, err := core.CopyFromRObject(slackTask)
	if err != nil {
		return err
	}
	_, _, err = s.datasource.Apply(ns, types.SlackTask, slackTask.GetName(), unstructuredData, false)
	return err
}

func (s *Service) ListSlackTask(ns string) ([]*v1.SlackTask, string, error) {
	unstructuredList, err := s.datasource.List(ns, types.SlackTask, "", 0, 0, nil)
	if err != nil {
		return nil, "", err
	}

	result := make([]*v1.SlackTask, 0)
	for _, unstructuredData := range unstructuredList.Items {
		slackTask := v1.SlackTask{}
		if err := core.CopyToRuntimeObject(&unstructuredData, &slackTask); err != nil {
			return nil, "", err
		}
		result = append(result, &slackTask)
	}
	return result, unstructuredList.GetResourceVersion(), nil
}

func (s *Service) GetFilter(ns, name string) (*v1.Filter, error) {
	unstructuredData, err := s.datasource.Get(ns, types.Filter, name)
	if err != nil {
		return nil, err
	}

	filter := &v1.Filter{}
	if err := core.CopyToRuntimeObject(unstructuredData, filter); err != nil {
		return nil, err
	}
	return filter, nil
}

func (s *Service) WatchSlackTask(ns, resourceVersion string) (<-chan watch.Event, error) {
	return s.datasource.Watch(ns, types.SlackTask, resourceVersion, 0, "")
}

func (s *Service) GetSlack(ns string) (*v1.Slack, error) {
	unstructuredData, err := s.datasource.Get(ns, types.Slack, fmt.Sprintf(common.NamespaceSlackName, ns))
	if err != nil {
		return nil, err
	}

	slack := &v1.Slack{}
	if err := core.CopyToRuntimeObject(unstructuredData, slack); err != nil {
		return nil, err
	}
	return slack, nil
}

func (s *Service) UpdateSlack(slack *v1.Slack) error {
	unstructuredData, err := core.CopyFromRObject(slack)
	if err != nil {
		return err
	}
	_, _, err = s.datasource.Apply(slack.Namespace, types.Slack, slack.Name, unstructuredData, false)
	return err
}

func (s *Service) WatchSlack(ns, resourceVersion string) (<-chan watch.Event, error) {
	return s.datasource.Watch(ns, types.Slack, resourceVersion, 0, nil)
}

func (s *Service) ListPod(ns string, selector string) ([]*corev1.Pod, string, error) {
	result := make([]*corev1.Pod, 0)
	unstructuredList, err := s.datasource.List(ns, types.Pod, "", 0, 0, selector)
	if err != nil {
		return nil, "", err
	}
	for _, unstructuredData := range unstructuredList.Items {
		pod := corev1.Pod{}
		if err := core.CopyToRuntimeObject(&unstructuredData, &pod); err != nil {
			return nil, "", err
		}
		result = append(result, &pod)
	}
	return result, unstructuredList.GetResourceVersion(), nil
}

func (s *Service) WatchPod(ns string, resourceVersion, selector string) (<-chan watch.Event, error) {
	return s.datasource.Watch(ns, types.Pod, resourceVersion, 0, selector)
}

func (s *Service) GetSink(ns string) (*v1.Sink, error) {
	unstructuredData, err := s.datasource.Get(ns, types.Sink, fmt.Sprintf(common.NamespaceSinkName, ns))
	if err != nil {
		return nil, err
	}

	sink := &v1.Sink{}
	if err := core.CopyToRuntimeObject(unstructuredData, sink); err != nil {
		return nil, err
	}
	return sink, nil
}

func (s *Service) ListTask(ns string) ([]core.Object, error) {
	result := make([]core.Object, 0)

	items, err := s.datasource.List(ns, types.Slack, "", 0, 0, nil)
	if err != nil {
		return nil, err
	}

	for _, item := range items.Items {
		result = append(result, core.FromUnstructured(item))
	}

	return result, nil
}

func NewIService(datasource datasource.IDataSource) IService {
	return &Service{datasource: datasource}
}
