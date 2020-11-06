package backends

import (
	"crypto/sha1"
	"encoding/base32"
	"fmt"
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
	log.Tracef("getting content for message %s from %s", e.hash, e.filename)
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
		return FilesystemMessageProvider{}, fmt.Errorf("Not a directory: %s", directory)
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

func (f FilesystemMessageProvider) LoadMessages() (messages []Message, err error) {
	log.Tracef("loading messages from %s", f.directory)
	entries, err := f.list()
	if err != nil {
		return nil, err
	}
	log.Debugf("found %d messages in %s", len(entries), f.directory)
	var res []Message
	for _, entry := range entries {
		content, err := entry.GetContent()
		if err != nil {
			log.Warnf("Could not read file %s - skipping it (%s)", entry.filename, err)
			break
		}
		log.Tracef("Found message %s (%s): size %d bytes", entry.hash, entry.filename, len(content))
		message := Message{
			id:      entry.hash,
			content: content,
		}
		res = append(res, message)
	}
	return res, nil
}

func (f FilesystemMessageProvider) DeleteMessage(messageId string) (err error) {
	entries, err := f.list()
	for _, entry := range entries {
		if entry.hash == messageId {
			log.Debugf("Deleting message %s (file %s)", entry.hash, entry.filename)
			return os.Remove(entry.filename)
		}
	}
	return fmt.Errorf("%s not found", err)
}
