package main

import (
	"flag"
	"github.com/DevelHell/popgun"
	"github.com/linkyard/pop3-webhook-server/internal/backends"
	log "github.com/sirupsen/logrus"
	"sync"
)

var user = flag.String("user", "mail", "Username that can be used to connect via POP3")
var password = flag.String("password", "", "Password that is used to connect via POP3 (required)")
var listenInterface = flag.String("interface", "localhost:1100", "Interface (host:port) to listen on")
var maildir = flag.String("mail-dir", "", "Directory to store the mail in (required unless --in-memory-only)")
var inmem = flag.String("in-memory-only", "false", "Set to true to not persist the messages to disk (default: false)")
var logLevel = flag.String("log-level", "Info", "Log level")

func main() {
	flag.Parse()
	parsedLogLevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("invalid log level '%s' falling back to default level 'Debug'", logLevel)
		parsedLogLevel = log.DebugLevel
	}
	log.SetLevel(parsedLogLevel)
	if len(*password) == 0 {
		log.Fatal("password not specified")
	}

	cfg := popgun.Config{
		ListenInterface: *listenInterface,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	log.Infof("Use the user '%s' to connect to %s using POP3", *user, *listenInterface)

	var messageProvider backends.MessageProvider
	if *inmem == "true" {
		log.Warn("Will not persist the messages, using in-memory mode")
		messageProvider = backends.NewInMemoryMessageProvider()
	} else {
		if len(*maildir) == 0 {
			log.Fatal("mail-dir not set")
		}
		log.Infof("Will persist the messages to %s", *maildir)
		messageProvider, err = backends.NewFilesystemMessageProvider(*maildir)
		if err != nil {
			log.Fatalf("could not start message provider: %s", err)
		}
	}

	backend := backends.NewSingleUserBackend(messageProvider, *user)
	authenticator := backends.NewSingleUserAuthenticator(*user, *password)

	server := popgun.NewServer(cfg, authenticator, backend)
	err = server.Start()
	if err != nil {
		log.Fatal(err)
		return
	}
	wg.Wait()
}
