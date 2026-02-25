package common

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// Log contains the default logger for the analyses service.
var Log = logrus.WithFields(logrus.Fields{
	"service": "analyses",
	"art-id":  "analyses",
	"group":   "org.cyverse",
})

// ErrorResponse represents an HTTP response body containing error information.
// Implements the error interface so it can be returned as an error.
type ErrorResponse struct {
	Message   string          `json:"message"`
	ErrorCode string          `json:"error_code,omitempty"`
	Details   *map[string]any `json:"details,omitempty"`
}

// ErrorBytes returns a byte-array representation of an ErrorResponse.
func (e ErrorResponse) ErrorBytes() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		Log.Errorf("unable to marshal %+v as JSON", e)
		return make([]byte, 0)
	}
	return bytes
}

// Error returns a string representation of an ErrorResponse.
func (e ErrorResponse) Error() string {
	return string(e.ErrorBytes())
}

// NewErrorResponse constructs an ErrorResponse from an error.
func NewErrorResponse(err error) ErrorResponse {
	switch val := err.(type) {
	case ErrorResponse:
		return val
	default:
		return ErrorResponse{Message: val.Error()}
	}
}

// NewErrorResponseWithCode constructs an ErrorResponse with an explicit error code.
func NewErrorResponseWithCode(err error, code string) ErrorResponse {
	return ErrorResponse{Message: err.Error(), ErrorCode: code}
}
