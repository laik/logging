package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yametech/logging/pkg/service"
	"io"
	"sync"
)

type Server struct {
	addr    string
	ns      string
	engine  *gin.Engine
	clients map[string]string
	lock    sync.Mutex
	cmdChan chan string
	service service.IService
}

func NewServer(addr string, ns string, service service.IService) *Server {
	engine := gin.Default()
	return &Server{
		addr:    addr,
		engine:  engine,
		ns:      ns,
		lock:    sync.Mutex{},
		service: service,
		clients: make(map[string]string),
	}
}

func (s *Server) addClient(node string, ip string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.clients[node] = ip
}

func (s *Server) removeClient(ip string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.clients, ip)
}

func (s *Server) Start() <-chan error {
	errors := make(chan error, 2)
	s.engine.GET("/:node", s.task)
	go func() {
		taskChan, err := s.service.WatchTask(s.ns)
		if err != nil {
			errors <- fmt.Errorf("service watch task chan error")
		}
		for {
			select {
			case task, ok := <-taskChan:
				if !ok {
					errors <- fmt.Errorf("task chan error")
					break
				}
				_ = task
				//TODO Task to Cmd
				s.cmdChan <- "task"
			}
		}

	}()
	go func() {
		errors <- s.engine.Run(s.addr)
	}()
	return errors

}

func (s *Server) task(g *gin.Context) {
	chanStream := make(chan string)
	closeChan := make(chan struct{})
	go func() {
		s.addClient(g.Param("node"), g.ClientIP())
		defer s.removeClient(g.Param("node"))
		for {
			select {
			case cmd, ok := <-s.cmdChan:
				if !ok {
					close(chanStream)
					break
				}
				chanStream <- cmd
			case <-closeChan:
				close(chanStream)
				break
			}
		}
	}()

	g.Stream(func(w io.Writer) bool {
		select {
		case item, ok := <-chanStream:
			if !ok {
				return false
			}
			g.SSEvent("", item)
		case <-g.Writer.CloseNotify():
			closeChan <- struct{}{}
			close(closeChan)
		}
		return true
	})
}
