package service

import (
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/datasource"
	"github.com/yametech/logging/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
)

type IService interface {
	WatchTask(ns string) (<-chan *v1.Slack, error)
	GetOutput(ns, name string) (*v1.Sink, error)
}

type Service struct {
	datasource datasource.IDataSource
}

func (s *Service) GetOutput(ns, name string) (*v1.Sink, error) {
	data, err := s.datasource.Get(ns, types.Sink, name)
	if err != nil {
		return nil, err
	}
	output := &v1.Sink{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(data.Object, output); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Service) WatchTask(ns string) (<-chan *v1.Slack, error) {
	result := make(chan *v1.Slack)
	tasks, err := s.datasource.List(ns, types.Slack, "", 0, 0, nil)
	if err != nil {
		return nil, err
	}
	//for _, task := range tasks.Items {
	//	result <- &task
	//}

	_ = tasks
	return result, nil
}

func NewIService(datasource datasource.IDataSource) IService {
	return &Service{datasource: datasource}
}
