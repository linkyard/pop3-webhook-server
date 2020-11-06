package backends

import log "github.com/sirupsen/logrus"

type InMemoryMessageProvider struct {
	messages map[string]Message
}

func NewInMemoryMessageProvider() InMemoryMessageProvider {
	return InMemoryMessageProvider{messages: map[string]Message{}}
}

func (p InMemoryMessageProvider) LoadMessages() (messages []Message, err error) {
	log.Debugf("Loading messages, %d found", len(p.messages))
	var res = make([]Message, len(p.messages))
	var i int = 0
	for _, msg := range p.messages {
		res[i] = msg
		i++
	}
	return res, nil
}

func (p InMemoryMessageProvider) DeleteMessage(messageId string) (err error) {
	log.Debugf("Deleted message %s", messageId)
	delete(p.messages, messageId)
	return nil
}

func (p InMemoryMessageProvider) AddMessage(id string, content string) {
	log.Debugf("Added message %s", id)
	p.messages[id] = Message{
		id:      id,
		content: content,
	}
}
