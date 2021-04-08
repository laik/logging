package command

import (
	"encoding/json"
	"testing"
)

func Test_run_command(t *testing.T) {
	cmd := NewCmd()
	pod1 := NewPod().SetName("echeor-api").SetNodeName("node1").SetOffset(1024).AddIp("127.0.0.1").AddIp("127.0.0.1")
	pod2 := NewPod().SetName("echeor-api").SetNodeName("node1").SetOffset(1023).AddIp("127.0.0.1").AddIp("128.0.0.1")
	cmd.SetOutput("kafka:test@10.200.100.200:9092").SetNs("test").SetFilter(1000, "").AddPod(pod1).AddPod(pod2).Run()

	cmdStr, err := cmd.ToString()
	if err != nil {
		t.Fatal(err)
	}
	expectedCmd := NewCmd()
	if err := json.Unmarshal([]byte(cmdStr), &expectedCmd); err != nil {
		t.Fatalf("%s", err)
	}

	if len(expectedCmd.Pods) != 1 || expectedCmd.Op != RUN || expectedCmd.Output != "kafka:test@10.200.100.200:9092" || expectedCmd.Ns != "test" || expectedCmd.Pods[0].Offset != 1023 {
		t.Fatal("unexpected test pod cmd")
	}
}
