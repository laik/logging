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
	"github.com/yametech/logging/pkg/utils"
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

func (s *Server) handle(evt <-chan watch.Event, errors chan error) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			cmdStr, _ := command.NewCmd().Hello().SetFilter("0", "").ToString()
			s.broadcast.Publish(cmdStr)

		case slackEvent, ok := <-evt:
			if !ok {
				errors <- fmt.Errorf("task chan error")
				break
			}

			slack := v1.Slack{}
			err := core.Convert(slackEvent.Object, &slack)
			if err != nil {
				fmt.Printf("[WARN] failed to get slack convert to core object %s\n", slackEvent.Object)
				continue
			}

			switch slackEvent.Type {
			case watch.Modified, watch.Added:
				// add tasks
				for index, task := range slack.Spec.AddTasks {
					sink, err := s.service.GetSink(task.Ns)
					if err != nil {
						fmt.Printf("[ERROR] failed get sink on slack task (%s) error: %s\n", task.ServiceName, err)
						continue
					}

					cmdStr, err := taskToCmd(command.RUN, &task, string(*sink.Spec.Type), *sink.Spec.Address)
					if err != nil {
						fmt.Printf("[ERROR] failed to convert to cmd string, data: (%v)\n", task)
						continue
					}

					exist, err := utils.CheckTopicExist(*sink.Spec.Address, task.ServiceName)
					if err != nil {
						fmt.Printf("[ERROR] check sink address (%s) topic (%s) error: (%v)\n", *sink.Spec.Address, task.ServiceName, err)
						continue
					}
					if !exist {
						if err := utils.CreateTopic(*sink.Spec.Address, task.ServiceName, *sink.Spec.Partition); err != nil {
							fmt.Printf("[ERROR] create topic address (%s) topic (%s) error: (%v)\n", *sink.Spec.Address, task.ServiceName, err)
							continue
						}
					}

					s.broadcast.Publish(cmdStr)

					slack.Spec.AddTasks = remove(slack.Spec.AddTasks, index)
				}

				// delete tasks
				for index, task := range slack.Spec.DeleteTasks {
					cmdStr, err := taskToCmd(command.STOP, &task, "", "")
					if err != nil {
						fmt.Printf("[WARN] failed to convert to cmd string %v\n", task)
						continue
					}

					s.broadcast.Publish(cmdStr)

					slack.Spec.DeleteTasks = remove(slack.Spec.DeleteTasks, index)
				}

			case watch.Deleted:
				for _, record := range slack.Spec.AllTasks {

					cmdStr, err := recordToCmd(command.STOP, &record)
					if err != nil {
						fmt.Printf("%s failed to convert to cmd string %v\n", common.WARN, record)
						continue
					}

					s.broadcast.Publish(cmdStr)
				}
			}

			if err := s.service.UpdateSlack(s.ns, slack.Name, &slack); err != nil {
				fmt.Printf("[ERROR] failed update slack after handle task (%v)", slack)
				continue
			}

			s.setResourceVersion(slack.ResourceVersion)
		}
	}
}

func (s *Server) watchSlack(errors chan error) {
	slackEventChan, err := s.service.WatchSlack(s.ns, s.slackResourceVersion)
	if err != nil {
		errors <- fmt.Errorf("service watch task chan error")
	}
	s.handle(slackEventChan, errors)
}

func (s *Server) firstCMDs() ([]string, *v1.Slack, error) {
	result := make([]string, 0)
	slack, err := s.service.GetSlack(s.ns, fmt.Sprintf(common.NamespaceSlackName, s.ns))
	if err != nil {
		fmt.Printf("%s loop collect client info , failed get slack (%s)\n", common.ERROR, fmt.Sprintf(common.NamespaceSlackName, s.ns))
		return nil, nil, err
	}

	for _, record := range slack.Spec.AllTasks {
		if !record.IsUpload {
			continue
		}
		cmdStr, err := recordToCmd(command.RUN, &record)
		if err != nil {
			fmt.Printf("[WARN] failed to convert to cmd string %v\n", record)
			continue
		}
		result = append(result, cmdStr)
	}

	for _, task := range slack.Spec.AddTasks {
		cmdStr, err := taskToCmd(command.STOP, &task, "", "")
		if err != nil {
			fmt.Printf("[WARN] failed to convert to cmd string %v\n", task)
			continue
		}
		result = append(result, cmdStr)
	}

	// delete tasks
	for index, task := range slack.Spec.DeleteTasks {
		cmdStr, err := taskToCmd(command.STOP, &task, "", "")
		if err != nil {
			fmt.Printf("[WARN] failed to convert to cmd string %v\n", task)
			continue
		}
		result = append(result, cmdStr)
		slack.Spec.DeleteTasks = remove(slack.Spec.DeleteTasks, index)
	}

	return result, slack, nil
}

func remove(slice []v1.Task, s int) []v1.Task {
	return append(slice[:s], slice[s+1:]...)
}

func (s *Server) every5SecondCollect() {
	for {
		time.Sleep(5 * time.Second)
		slack, err := s.service.GetSlack(s.ns, fmt.Sprintf(common.NamespaceSlackName, s.ns))
		if err != nil {
			fmt.Printf("%s loop collect client info , failed get slack (%s)\n", common.ERROR, fmt.Sprintf(common.NamespaceSlackName, s.ns))
			continue
		}

		var allTasks = make([]v1.Record, 0)
		for _, k := range s.broadcast.GetClientIPs() {
			resp, err := client.NewHttpClient().IP(clientIp(k)).Port("8080").Get("/pods")
			if err != nil {
				fmt.Printf("%s failed to get error %s\n", common.WARN, err)
				continue
			}
			var tmpRecords = make([]v1.Record, 0)
			if err := json.Unmarshal([]byte(resp), &tmpRecords); err != nil {
				fmt.Printf("%s failed to get node: (%s) unmarshal response data: (%s) error: (%s) \n", common.WARN, clientNode(k), resp, err)
				continue
			}
			for _, record := range tmpRecords {
				if record.IsUpload {
					allTasks = append(allTasks, record)
				}
			}
		}

		if len(allTasks) == 0 {
			continue
		}

		slack.Spec.AllTasks = allTasks

		if err := s.service.UpdateSlack(s.ns, slack.Name, slack); err != nil {
			fmt.Printf("%s failed update slack status, error: (%s) \n", common.ERROR, err)
		}
	}

}

func (s *Server) Start() <-chan error {
	errors := make(chan error, 2)
	s.engine.GET("/:node", s.task)

	go s.watchSlack(errors)
	go s.every5SecondCollect()

	go func() {
		errors <- s.engine.Run(s.addr)
	}()

	return errors

}

func (s *Server) task(g *gin.Context) {
	watchChannel := make(chan string, 0)

	id := clientID(g.Param("node"), g.ClientIP())
	s.broadcast.Registry(id, watchChannel)
	defer s.broadcast.UnRegistry(id)

	onceDo := false
	g.Stream(func(w io.Writer) bool {
		if !onceDo {
			cmds, slack, err := s.firstCMDs()
			if err != nil {
				fmt.Printf("[ERROR] failed list cmds")
				return false
			}
			if err := s.service.UpdateSlack(s.ns, slack.Name, slack); err != nil {
				fmt.Printf("[ERROR] failed update slack after first task (%v)", slack)
				return false
			}
			for _, cmd := range cmds {
				g.SSEvent("", cmd)
			}
			onceDo = true
		}

		select {
		case cmd, ok := <-watchChannel:
			if !ok {
				return false
			}
			g.SSEvent("", cmd)
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
	if strings.Contains(id, "::1") {
		return "127.0.0.1"
	}
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
