package http_tools

import (
	"github.com/educ-educ/gateway/internal/pkg/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ExcelError           string = "excel_error"
	FileServerError             = "file_server_error"
	DatabaseError               = "database_error"
	ValidationError             = "validation_error"
	AlreadyExist                = "already_exist"
	ForbiddenActionError        = "forbidden_action_error"
	NotFound                    = "not_found"
	ParseError                  = "parse_error"
	FileHeaderOpenError         = "file_header_open_error"
)

type Error struct {
	Type string `json:"type"`
	Info string `json:"info"`
}

func (err Error) Error() string {
	return "Type: " + err.Type + "; Info: " + err.Info
}

func (err Error) AsGinError() *gin.Error {
	return &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePrivate,
		Meta: nil,
	}
}

func ErrorsMiddleware(logger common.Logger, maxBodySizeToSave int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Waiting for all auth_handlers to execute
		c.Next()

		if len(c.Errors) < 1 {
			return
		}

		// We have at least one error
		code := http.StatusBadRequest

		errs := make([]error, 0, len(c.Errors))
		for _, err := range c.Errors {
			if customErr, ok := err.Err.(Error); ok {
				if customErr.Type == ExcelError || customErr.Type == FileServerError ||
					customErr.Type == DatabaseError || customErr.Type == FileHeaderOpenError {
					code = http.StatusInternalServerError
				}
				errs = append(errs, customErr)
				continue
			}
			errs = append(errs, err)
		}

		c.JSON(code, JSONErrors{Errors: errs})
	}
}
