package http_tools

import (
	"encoding/json"
	"net/http"
)

// JSONErrors
// @Description JSONErrors is a dto for cases one or more errors occurred during request handling
type JSONErrors struct {
	Errors []error `json:"errors"`
}

// JSONInfo
// @Description JSONInfo is a dto for cases request has no data to return.
// @Description JSONInfo is used to provide some information about what action happened
type JSONInfo struct {
	Info string `json:"info"`
}

func SendInfo(rw http.ResponseWriter, info string, code int) {
	rw.WriteHeader(code)
	jsonInfo := JSONInfo{Info: info}
	json.NewEncoder(rw).Encode(jsonInfo)
}
