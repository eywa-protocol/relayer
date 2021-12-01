package bridge

import "sync"

type QuitChannel chan struct{}

type QuitHandlers struct {
	mut      sync.Mutex
	handlers map[string]QuitChannel
	topics   map[QuitChannel]string
}

func NewQuitHandlers() *QuitHandlers {
	return &QuitHandlers{
		handlers: make(map[string]QuitChannel),
		topics:   make(map[QuitChannel]string),
	}
}

func (this *QuitHandlers) AddQuitHandler(topic string) QuitChannel {
	ch := make(QuitChannel)
	this.mut.Lock()
	defer this.mut.Unlock()
	if oldch, ok := this.handlers[topic]; ok {
		close(oldch)
		delete(this.topics, oldch)
	}
	this.handlers[topic] = ch
	this.topics[ch] = topic
	return ch
}

func (this *QuitHandlers) RemoveQuitHandler(ch QuitChannel) *string {
	this.mut.Lock()
	defer this.mut.Unlock()
	if topic, ok := this.topics[ch]; ok {
		if oldch, ok := this.handlers[topic]; ok && oldch == ch {
			close(oldch)
			delete(this.handlers, topic)
			delete(this.topics, ch)
			return &topic
		}
	}
	return nil
}

func (this *QuitHandlers) RemoveTopicHandler(topic string) bool {
	this.mut.Lock()
	defer this.mut.Unlock()
	if ch, ok := this.handlers[topic]; ok {
		close(ch)
		delete(this.handlers, topic)
		delete(this.topics, ch)
		return true
	}
	return false
}

func (this *QuitHandlers) DebugMapLength() (int, int) {
	return len(this.handlers), len(this.topics)
}
