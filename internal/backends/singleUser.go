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
	log.Infof("Authenticating user %s", user)
	if user == a.user && pass == a.password {
		log.Infof("Successfully authenticated user %s", user)
		return true
	} else {
		log.Warnf("Authentication for %s failed", user)
		return false
	}
}

type SingleUserBackend struct {
	user            string
	messageProvider MessageProvider
	session         *session
}
type session struct {
	locked   bool
	messages []Message
	deleted  map[string]bool
}

type Message struct {
	id      string
	content string
}
type MessageProvider interface {
	LoadMessages() (messages []Message, err error)
	DeleteMessage(messageId string) (err error)
}

func NewSingleUserBackend(messageProvider MessageProvider, user string) SingleUserBackend {
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
	log.Debugf("Stat")
	if !i.session.locked || i.session.messages == nil {
		return 0, 0, fmt.Errorf("not locked (stat)")
	}
	var count int = 0
	var size int = 0
	for _, msg := range i.session.messages {
		if !i.session.deleted[msg.id] {
			size += len(msg.content)
		}
	}
	return count, size, nil
}

// List of sizes of all messages in bytes (octets)
func (i SingleUserBackend) List(user string) (octets []int, err error) {
	log.Debugf("List")
	if !i.session.locked || i.session.messages == nil {
		return nil, fmt.Errorf("not locked (list)")
	}
	var res []int = make([]int, len(i.session.messages))
	for i, msg := range i.session.messages {
		res[i] = len(msg.content)
	}
	return res, nil
}

// Returns whether message exists and if yes, then return size of the message in bytes (octets)
func (i SingleUserBackend) ListMessage(user string, msgId int) (exists bool, octets int, err error) {
	log.Debugf("ListMessage %s", msgId)
	msg, err := i.getMessage(user, msgId, false)
	if err != nil {
		return false, 0, err
	}
	if msg != nil {
		return true, len(msg.content), nil
	} else {
		return false, 0, nil
	}

}

func (i SingleUserBackend) getMessage(user string, msgId int, failIfNotFound bool) (message *Message, err error) {
	if !i.session.locked || i.session.messages == nil {
		return nil, fmt.Errorf("not locked (getMessage)")
	}
	// TODO zero or 1 based?
	if msgId >= 0 && msgId < len(i.session.messages) {
		return &i.session.messages[msgId], nil
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
	log.Debugf("Retr %s", msgId)
	msg, err := i.getMessage(user, msgId, true)
	if err != nil {
		return "", nil
	}
	return msg.content, nil
}

// Delete message by message ID - message should be just marked as deleted until
// Update() is called. Be aware that after Dele() is called, functions like List() etc.
// should ignore all these messages even if Update() hasn't been called yet
func (i SingleUserBackend) Dele(user string, msgId int) error {
	log.Debugf("Dele %s", msgId)
	msg, err := i.getMessage(user, msgId, true)
	if err != nil {
		return err
	}
	i.session.deleted[msg.id] = true
	return nil
}

// Undelete all messages marked as deleted in single connection
func (i SingleUserBackend) Rset(user string) error {
	log.Debugf("Rset")
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
	log.Debugf("Uidl")
	if !i.session.locked || i.session.messages == nil {
		return nil, fmt.Errorf("not locked (uidl)")
	}
	var res []string = make([]string, len(i.session.messages))
	for i, msg := range i.session.messages {
		res[i] = msg.id
	}
	return res, nil
}

// Similar to ListMessage, but returns unique ID by message ID instead of size.
func (i SingleUserBackend) UidlMessage(user string, msgId int) (exists bool, uid string, err error) {
	log.Debugf("UidlMessage %s", msgId)
	msg, err := i.getMessage(user, msgId, false)
	if err != nil {
		return false, "", err
	}
	if msg != nil {
		return true, msg.id, nil
	} else {
		return false, "", nil
	}
}

// Write all changes to persistent storage, i.e. delete all messages marked as deleted.
func (i SingleUserBackend) Update(user string) (err error) {
	log.Debugf("Updating..")
	if !i.session.locked || i.session.messages == nil {
		return fmt.Errorf("not locked (update)")
	}
	err = nil
	var count int = 0
	for id, del := range i.session.deleted {
		if del {
			var err2 = i.messageProvider.DeleteMessage(id)
			count++
			if err2 != nil {
				log.Errorf("could not delete mail %s from storage: %s", id, err2)
				if err == nil {
					err = err2
				}
			}
		}
	}
	log.Infof("Update successful, deleted %d messages", count)
	return err
}

// Lock is called immediately after client is connected.
func (i SingleUserBackend) Lock(user string) error {
	log.Debugf("Locking...")
	if user != i.user {
		return fmt.Errorf("wrong user: %s (expected %s)", user, i.user)
	}
	if i.session.locked {
		return fmt.Errorf("already locked")
	}
	messages, err := i.messageProvider.LoadMessages()
	if err != nil {
		return err
	}
	i.session.locked = true
	i.session.messages = messages
	i.session.deleted = map[string]bool{}
	log.Debugf("Locked")
	return nil
}

// Release lock on storage, Unlock() is called after client is disconnected.
func (i SingleUserBackend) Unlock(user string) error {
	log.Debugf("Unlocked")
	if user == i.user {
		i.session.locked = false
		i.session.messages = nil
		i.session.deleted = nil
	}
	return nil
}
