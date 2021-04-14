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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
	"sync"
	"time"
)

type Server struct {
	addr string
	ns   string

	engine    *gin.Engine
	broadcast *Broadcast
	service   service.IService

	slackResourceVersion string
	podResourceVersion   string

	mutex    *sync.Mutex
	curSlack *v1.Slack

	needIgnoreBroadcast bool
}

func NewServer(addr string, ns string, service service.IService) *Server {
	return &Server{
		addr:  addr,
		ns:    ns,
		mutex: &sync.Mutex{},

		broadcast: NewBroadcast(),
		engine:    gin.Default(),
		service:   service,

		slackResourceVersion: "0",
		podResourceVersion:   "0",
		needIgnoreBroadcast:  false,
	}
}

func (p *Server) setNeedIgnoreBroadcastFalse() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.needIgnoreBroadcast = false
}

func (p *Server) podToTask(pod *corev1.Pod) v1.Task {
	pods := make([]v1.Pod, 0)
	for _, c := range pod.Spec.Containers {
		ips := make([]string, 0)
		for _, ip := range pod.Status.PodIPs {
			ips = append(ips, ip.IP)
		}
		pods = append(pods,
			v1.Pod{
				Node:      pod.Spec.NodeName,
				Pod:       pod.Name,
				Container: c.Name,
				Ips:       ips,
				Offset:    0,
			},
		)
	}
	return v1.Task{
		Ns:          pod.Namespace,
		ServiceName: pod.Labels["app"],
		Filter:      v1.Filter{},
		Pods:        pods,
	}
}

func (p *Server) slackAddNeedRunTasks(task v1.Task) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.curSlack == nil {
		slack, err := p.service.GetSlack(p.ns, fmt.Sprintf(common.NamespaceSlackName, p.ns))
		if err != nil {
			return err
		}
		p.curSlack = slack
	}

	if p.curSlack == nil {
		return nil
	}

	if p.curSlack.Spec.AddTasks == nil {
		p.curSlack.Spec.AddTasks = make([]v1.Task, 0)
		return nil
	}

	for _, _task := range p.curSlack.Spec.AddTasks {
		if _task.ServiceName == task.ServiceName && _task.Ns == task.Ns {
			continue
		}
	}

	p.curSlack.Spec.AddTasks = append(p.curSlack.Spec.AddTasks, task)

	return p.service.UpdateSlack(p.curSlack)
}

func (p *Server) slackAddNeedStopTasks(task v1.Task) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.curSlack == nil {
		slack, err := p.service.GetSlack(p.ns, fmt.Sprintf(common.NamespaceSlackName, p.ns))
		if err != nil {
			return err
		}
		p.curSlack = slack
	}

	if p.curSlack == nil {
		return nil
	}

	if p.curSlack.Spec.DeleteTasks == nil {
		p.curSlack.Spec.DeleteTasks = make([]v1.Task, 0)
		return nil
	}

	for _, _task := range p.curSlack.Spec.DeleteTasks {
		if _task.ServiceName == task.ServiceName && _task.Ns == task.Ns {
			continue
		}
	}

	p.curSlack.Spec.DeleteTasks = append(p.curSlack.Spec.DeleteTasks, task)

	return p.service.UpdateSlack(p.curSlack)
}

func (p *Server) podHandle(podEvt watch.Event) error {
	pod := corev1.Pod{}
	if err := core.Convert(podEvt.Object, &pod); err != nil {
		return err
	}

	if _, exist := pod.GetLabels()["app"]; !exist {
		return nil
	}

	switch podEvt.Type {
	case watch.Added, watch.Modified:
		if pod.Status.Phase == corev1.PodRunning {
			if err := p.slackAddNeedRunTasks(p.podToTask(&pod)); err != nil {
				return err
			}
		}
	case watch.Deleted:
		if err := p.slackAddNeedStopTasks(p.podToTask(&pod)); err != nil {
			return err
		}
	}
	return nil
}

func (p *Server) podWatchStart(errors chan error) {
	watchPodChan, err := p.service.WatchPod(p.ns, p.podResourceVersion, "")
	if err != nil {
		errors <- err
	}
	go func() {
		for {
			select {
			case evt, ok := <-watchPodChan:
				if !ok {
					errors <- fmt.Errorf("%s", "watch pod task chan error")
				}
				if err := p.podHandle(evt); err != nil {
					errors <- err
				}
			}
		}
	}()
}

func (p *Server) setSlackResourceVersion(version string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.slackResourceVersion = version
}

func (p *Server) setPodResourceVersion(version string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.slackResourceVersion = version
}

func (p *Server) slackHandle(evt <-chan watch.Event, errors chan error) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			cmdStr, _ := command.NewCmd().Hello().SetFilter("0", "").ToString()
			p.broadcast.Publish(cmdStr)

		case slackEvent, ok := <-evt:
			if !ok {
				errors <- fmt.Errorf("task chan error")
				break
			}

			slack := v1.Slack{}
			err := core.Convert(slackEvent.Object, &slack)
			if err != nil {
				fmt.Printf("%s failed to get slack convert to core object %s\n", common.WARN, slackEvent.Object)
				continue
			}

			switch slackEvent.Type {
			case watch.Modified, watch.Added:
				if p.needIgnoreBroadcast {
					p.setNeedIgnoreBroadcastFalse()
					continue
				}

				if err := p.broadcastRunTask(); err != nil {
					fmt.Printf("%s %s", common.WARN, err)
					continue
				}

				if err := p.broadcastStopTask(); err != nil {
					fmt.Printf("%s %s", common.WARN, err)
					continue
				}

			case watch.Deleted:
				for _, record := range slack.Spec.AllTasks {
					cmdStr, err := recordToCmd(command.STOP, &record)
					if err != nil {
						fmt.Printf("%s failed to convert to cmd string %v\n", common.WARN, record)
						continue
					}
					p.broadcast.Publish(cmdStr)
				}
			}

			p.setSlackResourceVersion(slack.ResourceVersion)
		}
	}
}

func (p *Server) broadcastRunTask() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.curSlack == nil {
		return nil
	}

	if p.curSlack.Spec.AddTasks == nil {
		return nil
	}

	for _, task := range p.curSlack.Spec.AddTasks {
		sink, err := p.service.GetSink(task.Ns)
		if err != nil {
			return fmt.Errorf("failed get sink on slack task (%s) error: %s\n", task.ServiceName, err)
		}

		cmdStr, err := taskToCmd(command.RUN, &task, string(*sink.Spec.Type), *sink.Spec.Address)
		if err != nil {
			return fmt.Errorf("failed to convert to cmd string, data: (%v)\n", task)
		}

		exist, err := utils.CheckTopicExist(*sink.Spec.Address, task.ServiceName)
		if err != nil {
			return fmt.Errorf("check sink address (%s) topic (%s) error: (%v)\n", *sink.Spec.Address, task.ServiceName, err)

		}
		if !exist {
			if err := utils.CreateTopic(*sink.Spec.Address, task.ServiceName, *sink.Spec.Partition); err != nil {
				return fmt.Errorf("create topic address (%s) topic (%s) error: (%v)\n", *sink.Spec.Address, task.ServiceName, err)
			}
		}

		p.broadcast.Publish(cmdStr)

		if err := p.removeRunTask(&task); err != nil {
			return err
		}
	}

	p.needIgnoreBroadcast = true

	return nil
}

func (p *Server) broadcastStopTask() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.curSlack == nil {
		return nil
	}

	if p.curSlack.Spec.DeleteTasks == nil {
		return nil
	}

	for _, task := range p.curSlack.Spec.DeleteTasks {
		cmdStr, err := taskToCmd(command.STOP, &task, "", "")
		if err != nil {
			return err
		}
		p.broadcast.Publish(cmdStr)
		if err := p.removeStopTask(&task); err != nil {
			return err
		}
	}

	p.needIgnoreBroadcast = true

	return nil
}

func (p *Server) watchSlack(errors chan error) {
	slackEventChan, err := p.service.WatchSlack(p.ns, p.slackResourceVersion)
	if err != nil {
		errors <- fmt.Errorf("service watch task chan error")
	}
	p.slackHandle(slackEventChan, errors)
}

func (p *Server) slack() (*v1.Slack, error) {
	p.mutex.Lock()
	p.mutex.Unlock()

	if p.curSlack == nil {
		slack, err := p.service.GetSlack(p.ns, fmt.Sprintf(common.NamespaceSlackName, p.ns))
		if err != nil {
			return nil, err
		}
		p.curSlack = slack
	}
	return p.curSlack, nil
}

func (p *Server) firstCMDs() ([]string, *v1.Slack, error) {
	result := make([]string, 0)

	slack, err := p.slack()
	if err != nil {
		return nil, nil, err
	}

	for _, record := range slack.Spec.AllTasks {
		if !record.IsUpload {
			continue
		}
		cmdStr, err := recordToCmd(command.RUN, &record)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, cmdStr)
	}

	for _, task := range slack.Spec.AddTasks {
		cmdStr, err := taskToCmd(command.RUN, &task, "", "")
		if err != nil {
			return nil, nil, err
		}
		result = append(result, cmdStr)
	}

	for _, task := range slack.Spec.DeleteTasks {
		cmdStr, err := taskToCmd(command.STOP, &task, "", "")
		if err != nil {
			return nil, nil, err
		}
		result = append(result, cmdStr)

		if err := p.removeStopTask(&task); err != nil {
			return nil, nil, err
		}
	}

	return result, slack, nil
}

func (p *Server) removeRunTask(task *v1.Task) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.curSlack == nil {
		return nil
	}

	if p.curSlack.Spec.AddTasks == nil {
		p.curSlack.Spec.AddTasks = make([]v1.Task, 0)
		return nil
	}

	for index, _task := range p.curSlack.Spec.AddTasks {
		if _task.Ns == task.Ns && _task.ServiceName == task.ServiceName {
			p.curSlack.Spec.AddTasks = append(
				p.curSlack.Spec.AddTasks[:index],
				p.curSlack.Spec.AddTasks[index+1:]...,
			)
		}
	}

	return p.service.UpdateSlack(p.curSlack)
}

func (p *Server) removeStopTask(task *v1.Task) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.curSlack == nil {
		return nil
	}

	if p.curSlack.Spec.DeleteTasks == nil {
		p.curSlack.Spec.DeleteTasks = make([]v1.Task, 0)
		return nil
	}

	for index, _task := range p.curSlack.Spec.DeleteTasks {
		if _task.Ns == task.Ns && _task.ServiceName == task.ServiceName {
			p.curSlack.Spec.DeleteTasks = append(
				p.curSlack.Spec.DeleteTasks[:index],
				p.curSlack.Spec.DeleteTasks[index+1:]...,
			)
		}
	}

	return p.service.UpdateSlack(p.curSlack)
}

func (p *Server) removeRecord(record *v1.Record) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.curSlack == nil {
		return nil
	}

	if p.curSlack.Spec.AllTasks == nil {
		p.curSlack.Spec.AllTasks = make([]v1.Record, 0)
		return nil
	}

	for index, _task := range p.curSlack.Spec.AllTasks {
		if _task.Ns == record.Ns && _task.ServiceName == record.ServiceName {
			p.curSlack.Spec.AllTasks = append(
				p.curSlack.Spec.AllTasks[:index],
				p.curSlack.Spec.AllTasks[index+1:]...,
			)
		}
	}

	return p.service.UpdateSlack(p.curSlack)
}

func (p *Server) every5SecondCollect() {
	for {
		time.Sleep(5 * time.Second)

		for _, k := range p.broadcast.GetClientIPs() {
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
					if err := p.addRecord(&record); err != nil {
						fmt.Printf("%s failed add record (%s) error: (%s) \n", common.WARN, record.ServiceName, err)
					}
					continue
				}

				if err := p.removeRecord(&record); err != nil {
					fmt.Printf("%s failed remove record (%s) error: (%s) \n", common.WARN, record.ServiceName, err)
					continue
				}
			}
		}
	}

}

func (p *Server) addRecord(record *v1.Record) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.curSlack == nil {
		return nil
	}

	if p.curSlack.Spec.AllTasks == nil {
		p.curSlack.Spec.AllTasks = make([]v1.Record, 0)
		return nil
	}
	p.curSlack.Spec.AllTasks = append(p.curSlack.Spec.AllTasks, *record)
	return p.service.UpdateSlack(p.curSlack)

}

func (p *Server) Start() <-chan error {
	errors := make(chan error, 2)
	p.engine.GET("/:node", p.task)

	go p.watchSlack(errors)
	go p.every5SecondCollect()
	go p.podWatchStart(errors)

	go func() {
		errors <- p.engine.Run(p.addr)
	}()

	return errors

}

func (p *Server) task(g *gin.Context) {
	watchChannel := make(chan string, 0)

	id := clientID(g.Param("node"), g.ClientIP())
	p.broadcast.Registry(id, watchChannel)
	defer p.broadcast.UnRegistry(id)

	onceDo := false
	g.Stream(func(w io.Writer) bool {
		if !onceDo {
			cmds, slack, err := p.firstCMDs()
			if err != nil {
				return false
			}

			if err := p.service.UpdateSlack(slack); err != nil {
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
