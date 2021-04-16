package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	stack "github.com/pkg/errors"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/command"
	"github.com/yametech/logging/pkg/common"
	"github.com/yametech/logging/pkg/core"
	"github.com/yametech/logging/pkg/service"
	"github.com/yametech/logging/pkg/utils"
	"io"
	"k8s.io/apimachinery/pkg/watch"
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
	s.engine.GET("/:node", s.task)

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
		NewSlack(ns, broadcast, service),
		//NewPod(ns, service),
		//NewSlackTask(ns, broadcast, service),
	}

	return &Server{
		addr:       addr,
		ns:         ns,
		engine:     gin.Default(),
		broadcast:  broadcast,
		service:    service,
		reconciles: reconciles,
	}
}

func getRecordsOffset(records []v1.Record, slackTask *v1.SlackTask) uint64 {
	for _, record := range records {
		if record.ServiceName == slackTask.Spec.ServiceName &&
			record.PodName == slackTask.Spec.Pod {
			return record.Offset
		}
	}
	return slackTask.Spec.Offset
}

func (s *Server) slackTaskToCMDStr(slackTask *v1.SlackTask) (string, error) {
	options := []command.Option{
		command.WithNs(s.ns),
		command.WithIPs(slackTask.Spec.Ips...),
		command.WithNodeName(slackTask.Spec.Node),
		command.WithPodName(slackTask.Spec.Pod),
		command.WithOffset(slackTask.Spec.Offset),
		command.WithServiceName(slackTask.Spec.ServiceName),
	}

	slack, err := s.service.GetSlack(s.ns)
	if err != nil {
		return "", stack.WithStack(err)
	}

	if slack.Spec.Records == nil {
		options = append(options, command.WithOffset(slackTask.Spec.Offset))
	} else {
		options = append(options, command.WithOffset(getRecordsOffset(slack.Spec.Records, slackTask)))
	}

	var filter *v1.Filter
	if slackTask.Spec.FilterName != "" {
		filter, err = s.service.GetFilter(s.ns, slackTask.Spec.FilterName)
		if err != nil {
			return "", stack.WithStack(err)
		}
	}
	if filter != nil {
		options = append(options, command.WithFilter(filter.Spec.MaxLength, filter.Spec.Expr))
	}

	sink, err := s.service.GetSink(s.ns)
	if err != nil {
		return "", stack.WithStack(err)
	}

	if sink == nil || sink.Spec.Address == nil || sink.Spec.Type == nil {
		fmt.Printf("%s sink not found or sink not define %v\n", common.WARN, sink)
		return "", nil
	}

	exist, err := utils.CheckTopicExist(*sink.Spec.Address, slackTask.Spec.ServiceName)
	if err != nil {
		return "", err
	}

	if !exist {
		if err := utils.CreateTopic(*sink.Spec.Address, slackTask.Spec.ServiceName, *sink.Spec.Partition); err != nil {
			return "", err
		}
	}

	options = append(options, command.WithOutput(fmt.Sprintf("%s:%s@%s", *sink.Spec.Type, slackTask.Spec.ServiceName, *sink.Spec.Address)))

	switch slackTask.Spec.Type {
	case watch.Added, watch.Modified:
		options = append(options, command.WithOp(command.RUN))
	case watch.Deleted:
		options = append(options, command.WithOp(command.STOP))
	}

	cmdStr, err := command.CMD(options...)
	if err != nil {
		return "", stack.WithStack(err)
	}

	return cmdStr, nil
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
		cmdStr, err := s.slackTaskToCMDStr(slackTask)
		if err != nil {
			return nil, stack.WithStack(err)
		}
		result = append(result, cmdStr)
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

			return true
		}

		select {
		case cmd, ok := <-watchChannel:
			if !ok {
				return false
			}
			exist, _node := core.GetByString(cmd, "node_name")
			if !exist || node != _node {
				return true
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
