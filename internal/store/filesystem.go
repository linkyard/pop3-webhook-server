package store

import (
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"github.com/linkyard/pop3-webhook-server/internal/backends"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FilesystemMessageProvider struct {
	directory string
}

type entry struct {
	hash     string
	filename string
	content  string
}

func (e entry) GetContent() (string, error) {
	log.Tracef("store: getting content for message %s from %s", e.hash, e.filename)
	content, err := ioutil.ReadFile(e.filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func NewFilesystemMessageProvider(directory string) (FilesystemMessageProvider, error) {
	dir, err := os.Stat(directory)
	if err != nil {
		return FilesystemMessageProvider{}, err
	}
	if !dir.IsDir() {
		return FilesystemMessageProvider{}, fmt.Errorf("not a directory: %s", directory)
	}
	return FilesystemMessageProvider{
		directory: directory,
	}, nil
}

func (f FilesystemMessageProvider) list() (entries []entry, err error) {
	files, err := ioutil.ReadDir(f.directory)
	if err != nil {
		return nil, err
	}
	var res []entry
	for _, file := range files {
		if file.IsDir() {
			break
		}
		hash := sha1.New()
		hash.Write([]byte(file.Name()))
		hashString := base32.StdEncoding.EncodeToString(hash.Sum(nil))
		message := entry{
			hash:     hashString,
			filename: filepath.Join(f.directory, file.Name()),
		}
		res = append(res, message)
	}
	return res, nil
}

func (f FilesystemMessageProvider) LoadMessages() (messages []backends.Message, err error) {
	log.Tracef("store: loading messages from %s", f.directory)
	entries, err := f.list()
	if err != nil {
		return nil, err
	}
	log.Debugf("store: found %d messages in %s", len(entries), f.directory)
	var res = make([]backends.Message, 0)
	for _, entry := range entries {
		content, err := entry.GetContent()
		if err != nil {
			log.Warnf("store: could not read file %s - skipping it (%s)", entry.filename, err)
			break
		}
		log.Tracef("store: found message %s (%s): size %d bytes", entry.hash, entry.filename, len(content))
		message := backends.Message{
			Id:      entry.hash,
			Content: content,
		}
		res = append(res, message)
	}
	return res, nil
}

func (f FilesystemMessageProvider) DeleteMessage(messageId string) (err error) {
	entries, err := f.list()
	for _, entry := range entries {
		if entry.hash == messageId {
			log.Debugf("store: deleting message %s (file %s)", entry.hash, entry.filename)
			return os.Remove(entry.filename)
		}
	}
	return fmt.Errorf("%s not found", err)
}

func (f FilesystemMessageProvider) StoreMessage(id string, content string) error {
	if len(id) == 0 {
		return fmt.Errorf("cannot store message: no id")
	}
	if  len(content) == 0 {
		return fmt.Errorf("cannot store message %s: no content", id)
	}
	log.Infof("store: Storing new message %s (%d bytes)", id, len(content))
	filename := filepath.Join(f.directory, id)
	return ioutil.WriteFile(filename, []byte(content), 0600)
}
