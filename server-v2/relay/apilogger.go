package relay

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Msger defines a type that uses a http response writer to relay messages.
type Msger interface {
	Init(w http.ResponseWriter)
	Msg(msg string)
}

// ErrMsger defines a type that uses a http response writer to relay
// errors based on a HTTP status code, as well as relaying messages.
type ErrMsger interface {
	Msger
	Err(msg string, statusCode int)
	StatusErr(statusCode int)
	ResErr(resBody any, msg string, statusCode int)
}

// APILogger is a means for the API endpoints to log messages. It implements
// ErrMsger.
type APILogger struct {
	w http.ResponseWriter
}

// NewAPILogger is the constructor for APILogger.
func NewAPILogger() *APILogger {
	return &APILogger{}
}

// Init initialises the http.ResponseWriter instance on the API(Err)Msger (eg.
// APILogger). This must be called before any other method of the logger. The
// ResponseWriter is initialised after NewAPILogger is called because we only have access
// to it inside the ServeHTTP methods of the HTTP handlers after we already
// initialised their dependencies.
func (l *APILogger) Init(w http.ResponseWriter) {
	l.w = w
}

// Msg relays an API message.
func (l *APILogger) Msg(msg string) {
	if _, err := l.w.Write([]byte(msg)); err != nil {
		l.Err(err.Error(), http.StatusInternalServerError)
	}
	log.Println(msg)
}

// Err logs an API error using a message text and HTTP status code.
func (l *APILogger) Err(msg string, statusCode int) {
	http.Error(l.w, msg, statusCode)
	log.Println(msg)
}

// StatusErr relays an API error based on a HTTP status code.
func (l *APILogger) StatusErr(statusCode int) {
	l.Err(http.StatusText(statusCode), statusCode)
}

// ResErr relays an API error based on an error response body.
func (l *APILogger) ResErr(resBody any, msg string, statusCode int) {
	l.w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(l.w).Encode(resBody); err != nil {
		msg := fmt.Sprintf("ERR: %s", err.Error())
		l.Err(msg, statusCode)
	}
	http.Error(l.w, msg, statusCode)
}