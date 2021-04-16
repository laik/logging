package api

import (
	"encoding/json"
	"fmt"
	stack "github.com/pkg/errors"
	v1 "github.com/yametech/logging/pkg/apis/yamecloud/v1"
	"github.com/yametech/logging/pkg/client"
	"github.com/yametech/logging/pkg/common"
	"github.com/yametech/logging/pkg/service"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

var _ IReconcile = &Slack{}

type Slack struct {
	ns        string
	broadcast *Broadcast
	service.IService
}

func NewSlack(ns string, broadcast *Broadcast, service service.IService) IReconcile {
	return &Slack{ns, broadcast, service}
}

func (s *Slack) loop(errors chan error) {
	var slack *v1.Slack
	var err error

RETRY:
	slack, err = s.GetSlack(s.ns)
	if err != nil || slack == nil {
		time.Sleep(5 * time.Second)
		goto RETRY
	}

	go func() {
		for {
			result := make([]v1.Record, 0)
			for _, id := range s.broadcast.GetClientIPs() {
				resp, err := client.NewHttpClient().IP(clientIp(id)).Port("8080").Get("/pods")
				if err != nil {
					fmt.Printf("%s failed to get error %s\n", common.WARN, err)
					continue
				}

				var tmp = make([]v1.Record, 0)
				if err := json.Unmarshal([]byte(resp), &tmp); err != nil {
					fmt.Printf("%s failed to get node: (%s) unmarshal response data: (%s) error: (%s) \n", common.WARN, clientNode(id), resp, err)
					continue
				}

				result = append(result, tmp...)
			}
			slack.Spec.Records = result

			if err := s.UpdateSlack(slack); err != nil {
				errors <- err
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func (s *Slack) Run(errors chan error) {
	//go s.loop(errors)

	var slack *v1.Slack
	var err error

	slack, err = s.GetSlack(s.ns)
	if err != nil {
		errors <- stack.WithStack(err)
		return
	}

	if slack == nil {
		if slack, err = s.CreateSlack(s.ns); err != nil {
			errors <- stack.WithStack(err)
			return
		}
	}

	slackChannel, err := s.WatchSlack(s.ns, "")
	if err != nil {
		errors <- stack.WithStack(err)
		return
	}

	for {
		slackEvt, _ := <-slackChannel

		switch slackEvt.Type {
		case watch.Error:
			errors <- fmt.Errorf("failed to watch slack %s", slackEvt.Object)
			return
		case watch.Deleted:
			if slack, err = s.CreateSlack(s.ns); err != nil {
				errors <- stack.WithStack(err)
				return
			}
		}
	}
}
