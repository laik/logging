package api

import (
	"fmt"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/command"
	"github.com/yametech/logging/pkg/core"
)

func taskList(o core.Object, path string) []v1.Task {
	result := make([]v1.Task, 0)
	taskMap, ok := o.Get(path).(map[string]v1.Task)
	if !ok {
		return result
	}

	for _, task := range taskMap {
		result = append(result, task)
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
