package webhook

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type ReceiveMailHandler struct {
	mailStore *MailStore
	token     string
}

type MailStore interface {
	StoreMessage(id string, content string) error
}

func NewReceiveMailHandler(token string, mailStore *MailStore) ReceiveMailHandler {
	return ReceiveMailHandler{
		token:     token,
		mailStore: mailStore,
	}
}

func (r ReceiveMailHandler) Register(router *mux.Router)  {
	router.HandleFunc("/", jsonResponse(index))
	router.HandleFunc("/health", jsonResponse(status))
	router.HandleFunc("/store", jsonResponse(r.store))
	router.PathPrefix("/").HandlerFunc(jsonResponse(notFound))
}

func index(_ *http.Request) (interface{}, httpError) {
	return indexResponse{Info: "Hello from pop3-webhook-server. Use POST on store to submit a message"}, nil
}

type indexResponse struct {
	Info string `json:"info"`
}

func status(_ *http.Request) (interface{}, httpError) {
	return statusResponse{Status: "ok"}, nil
}

type statusResponse struct {
	Status string `json:"status"`
}

func notFound(_ *http.Request) (interface{}, httpError) {
	return nil, clientError{
		detail:    "not found",
		errorCode: 404,
	}
}

func (r ReceiveMailHandler) store(req *http.Request) (interface{}, httpError) {
	log.Debug("Webhook request to store new mail")

	tokens, ok := req.URL.Query()["token"]
	if !ok || len(tokens) != 1 {
		return nil, clientError{errorCode: 401, detail: "missing token (query parameter)"}
	}
	if tokens[0] != r.token {
		return nil, clientError{errorCode: 403, detail: "token is not correct"}
	}

	request := storeRequest{}
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		return nil, clientError{errorCode: 400, detail: fmt.Sprintf("could not parse data: %s", err)}
	}
	if request.Id == 0 {
		return nil, clientError{errorCode: 400, detail: "missing id"}
	}
	if len(request.Body) == 0 {
		return nil, clientError{errorCode: 400, detail: "missing body"}
	}

	id := strconv.FormatInt(request.Id, 10)
	bodyBytes, err := base64.StdEncoding.DecodeString(request.Body)
	if err != nil {
		return nil, clientError{errorCode: 400, detail: fmt.Sprintf("error decoding body: %s", err)}
	}
	body := string(bodyBytes)

	err = (*r.mailStore).StoreMessage(id, body)
	if err != nil {
		return nil, serverError{detail: err.Error()}
	}
	return storeResponse{Id: id}, nil
}

type storeRequest struct {
	Id   int64  `json:"id"`
	Body string `json:"message"`
}

type storeResponse struct {
	Id string `json:"id"`
}
