package backends

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type SingleUserAuthenticator struct {
	user     string
	password string
}

func NewSingleUserAuthenticator(user string, password string) SingleUserAuthenticator {
	return SingleUserAuthenticator{
		user:     user,
		password: password,
	}
}

func (a SingleUserAuthenticator) Authorize(user, pass string) bool {
	log.Tracef("backend: authenticating user %s", user)
	if user == a.user && pass == a.password {
		log.Debugf("backend: successfully authenticated user %s", user)
		return true
	} else {
		log.Infof("backend: authentication for %s failed", user)
		return false
	}
}

type SingleUserBackend struct {
	user            string
	messageProvider *MessageProvider
	session         *session
}
type session struct {
	locked   bool
	messages []Message
	deleted  map[string]bool
}

type Message struct {
	Id      string
	Content string
}
type MessageProvider interface {
	LoadMessages() (messages []Message, err error)
	DeleteMessage(messageId string) (err error)
}

func NewSingleUserBackend(messageProvider *MessageProvider, user string) SingleUserBackend {
	session := session{}
	return SingleUserBackend{
		user:            user,
		messageProvider: messageProvider,
		session:         &session,
	}
}

// Returns total message count and total mailbox size in bytes (octets).
// Deleted messages are ignored.
func (i SingleUserBackend) Stat(user string) (messages, octets int, err error) {
	log.Debugf("backend: Stat")
	if user != i.user {
		return 0, 0, fmt.Errorf("wrong user: %s (expected %s)", user, i.user)
	}
	if !i.session.locked || i.session.messages == nil {
		return 0, 0, fmt.Errorf("not locked (stat)")
	}
	var count = 0
	var size = 0
	for _, msg := range i.session.messages {
		if !i.session.deleted[msg.Id] {
			size += len(msg.Content)
			count++
		}
	}
	log.Debugf("backend: Stat with %d messages (%d bytes)", count, size)
	return count, size, nil
}

// List of sizes of all messages in bytes (octets)
func (i SingleUserBackend) List(user string) (octets []int, err error) {
	log.Debugf("backend: List")
	if user != i.user {
		return nil, fmt.Errorf("wrong user: %s (expected %s)", user, i.user)
	}
	if !i.session.locked || i.session.messages == nil {
		return nil, fmt.Errorf("not locked (list)")
	}
	var res = make([]int, len(i.session.messages))
	for i, msg := range i.session.messages {
		res[i] = len(msg.Content)
	}
	return res, nil
}

// Returns whether message exists and if yes, then return size of the message in bytes (octets)
func (i SingleUserBackend) ListMessage(user string, msgId int) (exists bool, octets int, err error) {
	log.Debugf("backend: ListMessage %d", msgId)
	msg, err := i.getMessage(user, msgId, false)
	if err != nil {
		return false, 0, err
	}
	if msg != nil {
		return true, len(msg.Content), nil
	} else {
		return false, 0, nil
	}

}

func (i SingleUserBackend) getMessage(user string, msgId int, failIfNotFound bool) (message *Message, err error) {
	if user != i.user {
		return nil, fmt.Errorf("wrong user: %s (expected %s)", user, i.user)
	}
	if !i.session.locked || i.session.messages == nil {
		return nil, fmt.Errorf("not locked (getMessage)")
	}
	if msgId >= 1 && msgId <= len(i.session.messages) {
		return &i.session.messages[msgId-1], nil
	}
	if failIfNotFound {
		return nil, fmt.Errorf("ID %d was not found", msgId)
	} else {
		return nil, nil
	}
}

// Retrieve whole message by ID - note that message ID is a message position returned
// by List() function, so be sure to keep that order unchanged while client is connected
// See Lock() function for more details
func (i SingleUserBackend) Retr(user string, msgId int) (message string, err error) {
	log.Debugf("backend: Retr %d", msgId)
	msg, err := i.getMessage(user, msgId, true)
	if err != nil {
		return "", nil
	}
	return msg.Content, nil
}

// Delete message by message ID - message should be just marked as deleted until
// Update() is called. Be aware that after Dele() is called, functions like List() etc.
// should ignore all these messages even if Update() hasn't been called yet
func (i SingleUserBackend) Dele(user string, msgId int) error {
	log.Debugf("backend: Dele %d", msgId)
	msg, err := i.getMessage(user, msgId, true)
	if err != nil {
		return err
	}
	i.session.deleted[msg.Id] = true
	return nil
}

// Undelete all messages marked as deleted in single connection
func (i SingleUserBackend) Rset(user string) error {
	log.Debugf("backend: Rset")
	if user != i.user {
		return fmt.Errorf("wrong user: %s (expected %s)", user, i.user)
	}
	if !i.session.locked || i.session.messages == nil {
		return fmt.Errorf("not locked (rset)")
	}
	i.session.deleted = map[string]bool{}
	return nil
}

// List of unique IDs of all message, similar to List(), but instead of size there
// is a unique ID which persists the same across all connections. Uid (unique id) is
// used to allow client to be able to keep messages on the server.
func (i SingleUserBackend) Uidl(user string) (uids []string, err error) {
	log.Debugf("backend: Uidl")
	if user != i.user {
		return nil, fmt.Errorf("wrong user: %s (expected %s)", user, i.user)
	}
	if !i.session.locked || i.session.messages == nil {
		return nil, fmt.Errorf("not locked (uidl)")
	}
	var res = make([]string, len(i.session.messages))
	for i, msg := range i.session.messages {
		res[i] = msg.Id
	}
	return res, nil
}

// Similar to ListMessage, but returns unique ID by message ID instead of size.
func (i SingleUserBackend) UidlMessage(user string, msgId int) (exists bool, uid string, err error) {
	log.Debugf("backend: UidlMessage %d", msgId)
	msg, err := i.getMessage(user, msgId, false)
	if err != nil {
		return false, "", err
	}
	if msg != nil {
		return true, msg.Id, nil
	} else {
		return false, "", nil
	}
}

// Write all changes to persistent storage, i.e. delete all messages marked as deleted.
func (i SingleUserBackend) Update(user string) (err error) {
	log.Debugf("backend: Update")
	if user != i.user {
		return fmt.Errorf("wrong user: %s (expected %s)", user, i.user)
	}
	if !i.session.locked || i.session.messages == nil {
		return fmt.Errorf("not locked (update)")
	}
	err = nil
	var count = 0
	for id, del := range i.session.deleted {
		if del {
			var err2 = (*i.messageProvider).DeleteMessage(id)
			count++
			if err2 != nil {
				log.Errorf("backend: could not delete mail %s from storage: %s", id, err2)
				if err == nil {
					err = err2
				}
			}
		}
	}
	log.Infof("backend: update successful, deleted %d messages", count)
	return err
}

// Lock is called immediately after client is connected.
func (i SingleUserBackend) Lock(user string) error {
	log.Debugf("backend: Lock")
	if user != i.user {
		return fmt.Errorf("wrong user: %s (expected %s)", user, i.user)
	}
	if i.session.locked {
		return fmt.Errorf("already locked")
	}
	messages, err := (*i.messageProvider).LoadMessages()
	if err != nil {
		return err
	}
	i.session.locked = true
	i.session.messages = messages
	i.session.deleted = map[string]bool{}
	log.Debugf("backend: locked and loaded (%d messages)", len(i.session.messages))
	return nil
}

// Release lock on storage, Unlock() is called after client is disconnected.
func (i SingleUserBackend) Unlock(user string) error {
	log.Debugf("backend: Unlock")
	if user == i.user {
		i.session.locked = false
		i.session.messages = nil
		i.session.deleted = nil
	}
	return nil
}
