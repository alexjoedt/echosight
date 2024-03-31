package http

import "net/http"

type W map[string]any

type Response struct {
	Status  ResponseStatus `json:"status"`
	Code    string         `json:"code,omitempty"`
	Message string         `json:"message,omitempty"`
	Data    W              `json:"data,omitempty"`
	Errors  any            `json:"errors,omitempty"`
}

type ResponseStatus string

const (
	StatusOK    ResponseStatus = "ok"
	StatusWarn  ResponseStatus = "warning"
	StatusErr   ResponseStatus = "error"
	StatusFatal ResponseStatus = "fatal"
)

func (r *Response) WithMessage(msg string) *Response {
	r.Message = msg
	return r
}

func (r *Response) WithStatus(status ResponseStatus) *Response {
	r.Status = status
	return r
}

func (r *Response) WithErrors(errs any) *Response {
	r.Errors = errs
	return r
}

func OK(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusOK, &Response{
		Message: msg,
		Status:  StatusOK,
	})
}

func BadRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, &Response{
		Message: msg,
		Status:  StatusErr,
	})
}

func InternalServerError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, &Response{
		Message: msg,
		Status:  StatusErr,
	})
}

func TooManyRequestsError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusTooManyRequests, &Response{
		Message: msg,
		Status:  StatusErr,
	})
}

func Unauthorized(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusUnauthorized, &Response{
		Message: msg,
		Status:  StatusErr,
	})
}

func InvalidSession(w http.ResponseWriter) {
	w.Header().Set("Set-Cookie", "session")
	writeJSON(w, http.StatusUnauthorized, &Response{
		Message: "invalid session",
		Status:  StatusErr,
	})
}

func NotAuthenticaded(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, &Response{
		Message: "you must be authenticated to access this resource",
		Status:  StatusErr,
	})
}

func Forbidden(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusForbidden, &Response{
		Message: msg,
		Status:  StatusErr,
	})
}
