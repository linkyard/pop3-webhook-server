package store

import (
	"github.com/linkyard/pop3-webhook-server/internal/backends"
	log "github.com/sirupsen/logrus"
)

type InMemoryMessageProvider struct {
	messages map[string]backends.Message
}

func NewInMemoryMessageProvider() InMemoryMessageProvider {
	return InMemoryMessageProvider{messages: map[string]backends.Message{}}
}

func (p InMemoryMessageProvider) LoadMessages() (messages []backends.Message, err error) {
	log.Debugf("store: loading messages, %d found", len(p.messages))
	var res = make([]backends.Message, len(p.messages))
	var i = 0
	for _, msg := range p.messages {
		res[i] = msg
		i++
	}
	return res, nil
}

func (p InMemoryMessageProvider) DeleteMessage(messageId string) (err error) {
	log.Debugf("store: deleted message %s", messageId)
	delete(p.messages, messageId)
	return nil
}

func (p InMemoryMessageProvider) StoreMessage(id string, content string) error {
	log.Debugf("store: added message %s", id)
	p.messages[id] = backends.Message{
		Id:      id,
		Content: content,
	}
	return nil
}
