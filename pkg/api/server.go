package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/client"
	"github.com/yametech/logging/pkg/command"
	"github.com/yametech/logging/pkg/common"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/service"
	"io"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
	"time"
)

type Server struct {
	addr string
	ns   string

	engine    *gin.Engine
	broadcast *Broadcast
	service   service.IService

	slackResourceVersion string
}

func NewServer(addr string, ns string, service service.IService) *Server {
	return &Server{
		addr: addr,
		ns:   ns,

		broadcast: NewBroadcast(),
		engine:    gin.Default(),
		service:   service,

		slackResourceVersion: "0",
	}
}

func (s *Server) setResourceVersion(version string) {
	s.slackResourceVersion = version
}

func (s *Server) watchSlack(errors chan error) {
	slackEventChan, err := s.service.WatchSlack(s.ns, s.slackResourceVersion)
	if err != nil {
		errors <- fmt.Errorf("service watch task chan error")
	}

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			cmdStr, _ := command.NewCmd().Hello().ToString()
			s.broadcast.Publish(cmdStr)

		case slackEvent, ok := <-slackEventChan:
			if !ok {
				errors <- fmt.Errorf("task chan error")
				break
			}

			object, err := core.FromRuntimeObject(slackEvent.Object)
			if err != nil {
				fmt.Printf("[WARN] failed to get slack convert to core object %s\n", slackEvent.Object)
				continue
			}

			switch slackEvent.Type {

			case watch.Modified, watch.Added:
				// add tasks
				for _, task := range taskList(object, "spec.add_tasks") {
					sink, err := s.service.GetSink(task.Ns, task.Output)
					if err != nil {
						fmt.Printf("[WARN] failed to get sink on slack define output %v\n", task)
						continue
					}
					cmdStr, err := taskToCmd(command.RUN, &task, string(*sink.Spec.Type), *sink.Spec.Address)
					if err != nil {
						fmt.Printf("[WARN] failed to convert to cmd string %v\n", task)
						continue
					}
					s.broadcast.Publish(cmdStr)
				}

				// delete tasks
				for _, task := range taskList(object, "spec.delete_tasks") {
					cmdStr, err := taskToCmd(command.STOP, &task, "", "")
					if err != nil {
						fmt.Printf("[WARN] failed to convert to cmd string %v\n", task)
						continue
					}
					s.broadcast.Publish(cmdStr)
				}

			case watch.Deleted:
				podResults, ok := object.Get("status.all_tasks").([]v1.PodResult)
				if !ok {
					continue
				}

				for _, podResult := range podResults {
					task := v1.Task{
						Ns:     s.ns,
						Filter: v1.Filter{},
						Pods:   []v1.Pod{podResult.ToPod()},
					}
					cmdStr, err := taskToCmd(command.STOP, &task, "", "")
					if err != nil {
						fmt.Printf("%s failed to convert to cmd string %v\n", common.WARN, task)
						continue
					}
					s.broadcast.Publish(cmdStr)
				}
			}

			if version := object.Get("metadata.resourceVersion"); version != nil {
				s.setResourceVersion(version.(string))
			}
		}
	}
}

func (s *Server) scheduleTaskCollect() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			slack, err := s.service.GetSlack(s.ns, fmt.Sprintf(common.NamespaceSlackName, s.ns))
			if err != nil {
				fmt.Printf("%s failed get slack\n", common.ERROR)
				continue
			}

			allTasks := make([]v1.PodResult, 0)
			for _, k := range s.broadcast.GetClientIPs() {
				resp, err := client.NewHttpClient().IP(clientIp(k)).Port("8080").Get("/pods")
				if err != nil {
					fmt.Printf("%s failed to get node (%s) pods\n", common.WARN, clientNode(k))
					continue
				}
				podResult := v1.PodResult{}
				if err := json.Unmarshal([]byte(resp), &podResult); err != nil {
					fmt.Printf("%s failed to get node (%s) unmarshal response data\n", common.WARN, clientNode(k))
					continue
				}
				allTasks = append(allTasks, podResult)
			}

			if len(allTasks) == 0 {
				continue
			}

			slack.Status.AllTasks = allTasks

			if err := s.service.UpdateSlackStatus(s.ns, slack); err != nil {
				fmt.Printf("%s failed update slack status\n", common.ERROR)
			}
		}
	}

}

func (s *Server) Start() <-chan error {
	errors := make(chan error, 2)
	s.engine.GET("/:node", s.task)

	go s.watchSlack(errors)

	go func() {
		errors <- s.engine.Run(s.addr)
	}()

	go func() { s.scheduleTaskCollect() }()

	return errors

}

func (s *Server) task(g *gin.Context) {
	chanStream := make(chan string)
	id := clientID(g.Param("node"), g.ClientIP())
	s.broadcast.Registry(id, chanStream)
	defer s.broadcast.UnRegistry(id)

	g.Stream(func(w io.Writer) bool {
		select {
		case item, ok := <-chanStream:
			if !ok {
				return false
			}
			g.SSEvent("", item)
			//fmt.Printf("%s send to client %s msg: %v\n", common.INFO, id, item)
		case <-g.Writer.CloseNotify():
			return false
		}
		return true
	})
}

func clientID(node, ip string) string {
	return fmt.Sprintf("%s-%s", node, ip)
}

func clientIp(id string) string {
	if res := strings.Split(id, "-"); len(res) == 2 {
		return res[1]
	}
	return ""
}

func clientNode(id string) string {
	if res := strings.Split(id, "-"); len(res) == 2 {
		return res[0]
	}
	return ""
}
