package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yametech/logging/pkg/command"
	"github.com/yametech/logging/pkg/common"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/service"
	"io"
	"k8s.io/apimachinery/pkg/watch"
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

		broadcast: &Broadcast{make(map[string]chan string)},
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
				for _, task := range taskList(object, "status.all_tasks") {
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

func (s *Server) Start() <-chan error {
	errors := make(chan error, 2)
	s.engine.GET("/:node", s.task)

	go s.watchSlack(errors)

	go func() {
		errors <- s.engine.Run(s.addr)
	}()

	return errors

}

func (s *Server) task(g *gin.Context) {
	chanStream := make(chan string)
	id := fmt.Sprintf("%s-%s", g.Param("node"), g.ClientIP())
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
