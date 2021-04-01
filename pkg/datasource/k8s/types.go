package k8s

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

type InitResource func(*Resources)

type ResourceLister interface {
	Ranges(d dynamicinformer.DynamicSharedInformerFactory, stop <-chan struct{})
	GetGvr(string) (schema.GroupVersionResource, error)
}

type Resources struct {
	excluded []string

	Data map[string]schema.GroupVersionResource
}

func NewResources(excluded []string, inits ...InitResource) ResourceLister {
	rs := &Resources{
		excluded: excluded,
		Data:     make(map[string]schema.GroupVersionResource),
	}
	for _, init := range inits {
		init(rs)
	}
	return rs
}

func (m *Resources) Registry(s string, resource schema.GroupVersionResource) {
	if _, exist := m.Data[s]; exist {
		return
	}
	m.Data[s] = resource
}

func (m *Resources) Ranges(d dynamicinformer.DynamicSharedInformerFactory, stop <-chan struct{}) {
	for _, v := range m.excluded {
		value := v
		delete(m.Data, value)
	}
	for _, v := range m.Data {
		value := v
		go d.ForResource(value).Informer().Run(stop)
	}
}

func (m *Resources) GetGvr(s string) (schema.GroupVersionResource, error) {
	item, exist := m.Data[s]
	if !exist {
		return schema.GroupVersionResource{}, fmt.Errorf("resource (%s) not exist", s)
	}
	return item, nil
}
