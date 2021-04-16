package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	stack "github.com/pkg/errors"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/service"
	"io"
	"net/http"
	"strings"
)

type IReconcile interface {
	Run(chan error)
}

type Server struct {
	addr string
	ns   string

	engine    *gin.Engine
	broadcast *Broadcast
	service   service.IService

	taskCmdList map[string]string //serviceName cmdStr
	reconciles  []IReconcile
}

func (s *Server) Start() error {
	errors := make(chan error)
	go func() { errors <- s.engine.Run(s.addr) }()

	go func() {
		for _, reconcile := range s.reconciles {
			go reconcile.Run(errors)
		}
	}()

	return <-errors
}

func NewServer(addr string, ns string, service service.IService) *Server {
	broadcast := NewBroadcast()
	reconciles := []IReconcile{
		NewSlackTask(ns, broadcast, service),
		NewPod(ns, service),
		NewSlack(ns, service),
	}

	return &Server{
		addr:       addr,
		ns:         ns,
		engine:     gin.Default(),
		service:    service,
		reconciles: reconciles,
	}
}

func (s *Server) slackTaskToCMDStr(slackTask *v1.SlackTask) string {
	return ""
}

func (s *Server) getCMDsByNode(node string) ([]string, error) {
	result := make([]string, 0)

	slackTasks, _, err := s.service.ListSlackTask(s.ns)
	if err != nil {
		return nil, stack.WithStack(err)
	}

	for _, slackTask := range slackTasks {
		if slackTask.Spec.Node != node {
			continue
		}
		result = append(result, s.slackTaskToCMDStr(slackTask))
	}

	return result, nil
}

func (s *Server) task(g *gin.Context) {
	watchChannel := make(chan string, 0)

	node := g.Param("node")
	id := clientID(node, g.ClientIP())
	s.broadcast.Registry(id, watchChannel)
	defer s.broadcast.UnRegistry(id)

	onceDo := false
	g.Stream(func(w io.Writer) bool {
		if !onceDo {
			cmdList, err := s.getCMDsByNode(node)
			if err != nil {
				g.JSON(http.StatusInternalServerError, err)
				return true
			}
			for _, cmd := range cmdList {
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

func clientID(node, ip string) string { return fmt.Sprintf("%s-%s", node, ip) }

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
