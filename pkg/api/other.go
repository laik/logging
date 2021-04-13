package api

import (
	"encoding/json"
	"fmt"

	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/command"
	"github.com/yametech/logging/pkg/core"
)

func objectToTasks(object interface{}, task *v1.Task) error {
	bs, err := json.Marshal(object)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, task); err != nil {
		return err
	}
	return nil
}

func taskList(o core.Object, path string) []v1.Task {
	taskList := o.Get(path)

	result := make([]v1.Task, 0)
	if _taskList, ok := taskList.([]interface{}); ok {
		for _, obj := range _taskList {
			task := v1.Task{}
			if err := objectToTasks(obj, &task); err != nil {
				fmt.Printf("objectToTasks error: %s\n", err)
				continue
			}
			result = append(result, task)
		}
		return result
	}

	return result
}

func taskToCmd(op command.Op, task *v1.Task, outputType, address string) (string, error) {
	cmd := command.NewCmd()

	for _, pod := range task.Pods {
		cmd.AddPod(
			command.NewPod().
				SetIPs(pod.Ips...).
				SetName(pod.Pod).
				SetOffset(uint64(pod.Offset)).
				SetNodeName(pod.Node),
		)
	}
	cmd.SetNs(task.Ns).
		SetFilter(task.Filter.MaxLength, task.Filter.Expr).
		SetOutput(fmt.Sprintf("%s:%s@%s", outputType, task.ServiceName, address)).
		SetServiceName(task.ServiceName).
		SetOp(op)

	cmdStr, err := cmd.ToString()
	if err != nil {
		return "", err
	}
	return cmdStr, nil
}

func recordToCmd(op command.Op, record *v1.Record) (string, error) {
	cmd := command.NewCmd()

	pod := command.NewPod().
		SetIPs(record.Ips...).
		SetName(record.PodName).
		SetOffset(uint64(record.Offset)).
		SetNodeName(record.NodeName)

	cmd.AddPod(pod).SetNs(record.Ns).
		SetFilter(record.Filter.MaxLength, record.Filter.Expr).
		SetOutput(record.Output).
		SetServiceName(record.ServiceName).
		SetOp(op)

	cmdStr, err := cmd.ToString()
	if err != nil {
		return "", err
	}
	return cmdStr, nil
}
