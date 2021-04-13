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
	GetSlack(ns, name string) (*v1.Slack, error)
	UpdateSlack(ns, name string, slack *v1.Slack) error

	ListPod(ns string, selector string) ([]corev1.Pod, error)
	WatchPod(ns string, resourceVersion, selector string) (<-chan watch.Event, error)

	GetSink(ns string) (*v1.Sink, error)
}

type Service struct {
	datasource datasource.IDataSource
}

func (s *Service) GetSlack(ns, name string) (*v1.Slack, error) {
	unstructuredData, err := s.datasource.Get(ns, types.Slack, name)
	if err != nil {
		return nil, err
	}

	slack := &v1.Slack{}
	if err := core.CopyToRuntimeObject(unstructuredData, slack); err != nil {
		return nil, err
	}
	return slack, nil
}

func (s *Service) UpdateSlack(ns, name string, slack *v1.Slack) error {
	unstructuredData, err := core.CopyFromRObject(slack)
	if err != nil {
		return err
	}
	_, _, err = s.datasource.Apply(ns, types.Slack, name, unstructuredData, false)
	return err
}

func (s *Service) WatchSlack(ns, resourceVersion string) (<-chan watch.Event, error) {
	return s.datasource.Watch(ns, types.Slack, resourceVersion, 0, nil)
}

func (s *Service) ListPod(ns string, selector string) ([]corev1.Pod, error) {
	result := make([]corev1.Pod, 0)
	unstructuredList, err := s.datasource.List(ns, types.Pod, "", 0, 0, selector)
	if err != nil {
		return nil, err
	}
	for _, unstructuredData := range unstructuredList.Items {
		pod := corev1.Pod{}
		if err := core.CopyToRuntimeObject(&unstructuredData, &pod); err != nil {
			return nil, err
		}
		result = append(result, pod)
	}
	return result, nil
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
