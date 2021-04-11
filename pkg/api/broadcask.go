package api

import "sync"

type Broadcast struct {
	*sync.Mutex
	subscribes map[string]chan string
}

func NewBroadcast() *Broadcast {
	return &Broadcast{
		Mutex:      &sync.Mutex{},
		subscribes: make(map[string]chan string),
	}
}

func (b *Broadcast) GetClientIPs() []string {
	b.Lock()
	defer b.Unlock()
	result := make([]string, 0)
	for k, _ := range b.subscribes {
		result = append(result, k)
	}
	return result
}

func (b *Broadcast) Registry(name string, channel chan string) {
	b.Lock()
	defer b.Unlock()
	b.subscribes[name] = channel
}

func (b *Broadcast) UnRegistry(name string) {
	b.Lock()
	defer b.Unlock()
	if _, exist := b.subscribes[name]; !exist {
		return
	}
	delete(b.subscribes, name)
}

func (b *Broadcast) Publish(msg string) {
	b.Lock()
	defer b.Unlock()
	for _, subscribe := range b.subscribes {
		subscribe <- msg
	}
}
