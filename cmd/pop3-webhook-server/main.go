package main

import (
	"github.com/namsral/flag"
	"github.com/DevelHell/popgun"
	"github.com/linkyard/pop3-webhook-server/internal/backends"
	messageProvider2 "github.com/linkyard/pop3-webhook-server/internal/store"
	"github.com/linkyard/pop3-webhook-server/internal/webhook"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
)

var user = flag.String("user", "mail", "Username that can be used to connect via POP3")
var password = flag.String("password", "", "Password that is used to connect via POP3 (required)")
var token = flag.String("token", "", "Token that is required to submit via webhook (required)")
var httpAddr = flag.String("http-interface", "localhost:8080", "The address to listen on for HTTP requests.")
var popAddr = flag.String("pop-interface", "localhost:1100", "Interface (host:port) to listen on")
var maildir = flag.String("mail-dir", "", "Directory to store the mail in (required unless --in-memory-only)")
var inmem = flag.String("in-memory-only", "false", "Set to true to not persist the messages to disk (default: false)")
var logLevel = flag.String("log-level", "Info", "Log level (ERROR, WARN, INFO, DEBUG, TRACE)")

func main() {
	flag.Parse()
	parsedLogLevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("invalid log level '%s' falling back to default level 'Debug'", *logLevel)
		parsedLogLevel = log.DebugLevel
	}
	log.SetLevel(parsedLogLevel)
	if len(*password) == 0 {
		log.Fatal("password not specified")
	}
	if len(*token) == 0 {
		log.Fatal("token not specified")
	}

	cfg := popgun.Config{
		ListenInterface: *popAddr,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	log.Infof("Use the user '%s' to connect to %s using POP3", *user, *popAddr)

	var messageProvider backends.MessageProvider
	var mailStore webhook.MailStore
	if *inmem == "true" {
		log.Warn("Will not persist the messages, using in-memory mode")
		store := messageProvider2.NewInMemoryMessageProvider()
		messageProvider = store
		mailStore = store
	} else {
		if len(*maildir) == 0 {
			log.Fatal("mail-dir not set")
		}
		log.Infof("Will persist the messages to %s", *maildir)
		store, err := messageProvider2.NewFilesystemMessageProvider(*maildir)
		if err != nil {
			log.Fatalf("could not start message provider: %s", err)
		}
		messageProvider = store
		mailStore = store
	}

	backend := backends.NewSingleUserBackend(&messageProvider, *user)
	authenticator := backends.NewSingleUserAuthenticator(*user, *password)

	server := popgun.NewServer(cfg, authenticator, backend)
	err = server.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	httpHandler := webhook.NewReceiveMailHandler(*token, &mailStore)
	log.Infof("Started http server on %s (use /mail/v1/store)", *httpAddr)
	err = http.ListenAndServe(*httpAddr, httpHandler.Handler())
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
