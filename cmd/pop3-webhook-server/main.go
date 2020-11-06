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
var logLevel = flag.String("log-level", "Debug", "Log level")

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

	messageProvider := backends.NewInMemoryMessageProvider()
	messageProvider.AddMessage("11", "hi there")

	backend := backends.NewSingleUserBackend(messageProvider, *user)
	authorizator := backends.NewSingleUserAuthenticator(*user, *password)

	server := popgun.NewServer(cfg, authorizator, backend)
	err = server.Start()
	if err != nil {
		log.Fatal(err)
		return
	}
	wg.Wait()
}
