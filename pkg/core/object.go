package core

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

type Object map[string]interface{}

func (o Object) Set(path string, value interface{}) { Set(o, path, value) }

func (o Object) Get(path string) interface{} { return Get(o, path) }

func (o Object) Delete(path string) { Delete(o, path) }

func FromUnstructured(u unstructured.Unstructured) Object {
	return u.Object
}

func FromRuntimeObject(r runtime.Object) (Object, error) {
	bs, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	o := make(Object)
	if err := json.Unmarshal(bs, &o); err != nil {
		return nil, err
	}
	return o, nil
}

func CopyToRuntimeObject(src *unstructured.Unstructured, target runtime.Object) error {
	bytesData, err := json.Marshal(src.Object)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytesData, target)
}

func CopyFromRObject(src runtime.Object) (*unstructured.Unstructured, error) {
	bytesData, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	target := &unstructured.Unstructured{}
	if err := json.Unmarshal(bytesData, target); err != nil {
		return nil, err
	}
	return target, nil
}

func Convert(src runtime.Object, target runtime.Object) error {
	bytesData, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytesData, target); err != nil {
		return err
	}
	return nil
}
