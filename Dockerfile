FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git build-base

ENV USER=appuser
ENV UID=3050
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/app
COPY . .

RUN go get -d -v ./...
RUN go test -v ./...
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /pop3-webhook-server ./cmd/pop3-webhook-server/main.go


FROM alpine

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /pop3-webhook-server /usr/local/bin/pop3-webhook-server
RUN ls -l /usr/local/bin

RUN apk add --no-cache tini

ENTRYPOINT ["tini", "-g", "--"]

USER appuser:appuser
CMD ["/usr/local/bin/pop3-webhook-server"]
