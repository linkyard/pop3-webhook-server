![Go](https://github.com/linkyard/pop3-webhook-server/workflows/Go/badge.svg)
![Docker](https://github.com/linkyard/pop3-webhook-server/workflows/Docker/badge.svg)

# POP3 Server with Webhook

Simple POP3 server that gets its messages via a webhook. Single account/mailbox only.

Can be used to connect mail-based applications to a server such as [postal](https://postal.atech.media).


## Features
* Storage of mail in the filesystem (or in-memory)
* POP3 compliant
* simple user authentication (user+password, single user)
* webhook to receive mail
* simple authentication for the webhook (token)

## Usage
### Running
I recommend using the docker image: `ghcr.io/linkyard/pop3-webhook-server:latest` or  `eu.gcr.io/linkyard-191615/linkyard-cloud/pop3-webhook-server` (private)

Run by calling the executable with the following parameters (env or command line)
- `user`: Username that can be used to connect via POP3. Default: `mail`
- `password`: Password that is used to connect via POP3 (required)
- `token`: Token that is required to submit via webhook (required)")
- `http-interface`: The address to listen on for HTTP requests. Default: `localhost:8080`
- `base-path`: The base HTTP path the webhook should run under. Default: `/`
- `pop-interface`: Interface (host:port) to listen on. Default: `localhost:1100`
- `mail-dir`: Directory to store the mail in. Required unless in-memory-only is set
- `in-memory-only`: Set to `true` to not persist the messages to disk. Default: `false`
- `log-level`: Log level (ERROR, WARN, INFO, DEBUG, TRACE). Default: INFO

### Storing Mail
Store the mail by calling the webhook (default is `http://localhost:8080/store`).

Query Paramters:
- `token`: set to the same value as the startup parameter (used for authentication)

Body - application/json with the following fields:
- `id` (number): ID of the message
- `message` (string, base64 encoded): The raw content of the message

### Receiving Mails
Connect to the POP3 server at `localhost:1100` (or what you specified in pop-interface).
Use basic authentication with the user/password you specified at startup.

## Credits
* Uses https://github.com/DevelHell/popgun/ for POP3

## LICENSE
MIT, see [License](LICENSE)
