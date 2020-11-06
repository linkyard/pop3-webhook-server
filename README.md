# POP3 Server with Webhook

Simple POP3 server that gets its messages via a webhook. Single account/mailbox only.

Can be used to connect mail-based applications to a server such as [postal](https://postal.atech.media).


## Features
* Storage of mail in the filesystem (or in-memory)
* POP3 compliant
* simple user authentication (user+password, single user)
* webhook to receive mail
* simple authentication for the webhook (token)

## Credit
* Uses https://github.com/DevelHell/popgun/ for POP3

## LICENSE
MIT, see [License](LICENSE)
