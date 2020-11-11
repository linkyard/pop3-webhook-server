package webhook

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type errorResponse struct {
	Error  bool   `json:"error"`
	Code   int    `json:"code"`
	Detail string `json:"detail"`
}

type httpError interface {
	ErrorCode() int
	Detail() string
}

type serverError struct {
	detail string
}

func (s serverError) ErrorCode() int {
	return 500
}

func (s serverError) Detail() string {
	return s.detail
}

type clientError struct {
	detail    string
	errorCode int
}

func (s clientError) ErrorCode() int {
	return s.errorCode
}

func (s clientError) Detail() string {
	return s.detail
}

func jsonResponse(f func(request *http.Request) (interface{}, httpError)) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		log.Tracef("Handling request to %s", req.RequestURI)
		result, err := f(req)
		var writeErr error
		if err == nil {
			log.Debugf("Request to %s was successful", req.RequestURI)
			resp.Header().Set("Content-Type", "application/json")
			resp.WriteHeader(200)
			writeErr = json.NewEncoder(resp).Encode(result)
		} else {
			log.Infof("Request to %s failed: %d - %s", req.RequestURI, err.ErrorCode(), err.Detail())
			result := errorResponse{
				Error:  true,
				Code:   err.ErrorCode(),
				Detail: err.Detail(),
			}
			resp.Header().Set("Content-Type", "application/json")
			resp.WriteHeader(err.ErrorCode())
			writeErr = json.NewEncoder(resp).Encode(result)
		}
		if writeErr != nil {
			log.Debugf("Failed to write response to client %s", req.RemoteAddr)
		}
	}
}
