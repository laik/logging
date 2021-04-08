package api

type Broadcast struct {
	subscribes map[string]chan string
}

func (b *Broadcast) Registry(name string, channel chan string) {
	b.subscribes[name] = channel
}

func (b *Broadcast) UnRegistry(name string) {
	if _, exist := b.subscribes[name]; !exist {
		return
	}
	delete(b.subscribes, name)
}

func (b *Broadcast) Publish(msg string) {
	for _, subscribe := range b.subscribes {
		subscribe <- msg
	}
}
