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

func (r ReceiveMailHandler) Handler() *mux.Router {
	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/store", jsonResponse(r.store))
	return router
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
