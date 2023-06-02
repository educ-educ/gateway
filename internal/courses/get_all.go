package courses

import (
	"encoding/json"
	"github.com/educ-educ/gateway/internal/pkg/common"
	"github.com/gin-gonic/gin"
	"net/http"
)

type courseStruct struct {
	ID     int        `json:"id,omitempty"`
	Author string     `json:"author"`
	Name   string     `json:"name"`
	Body   bodyStruct `json:"body"`
}

type bodyStruct struct {
	Data string `json:"data"`
	Src  string `json:"src,omitempty"`
}

type outCourseStruct struct {
	ID   int    `json:"id"`
	Src  string `json:"src"`
	Text string `json:"text"`
	Data string `json:"data"`
}

type GetAllHandler struct {
	logger common.Logger
	url    string
}

func NewGetAllHandler(logger common.Logger, url string) *GetAllHandler {
	return &GetAllHandler{
		logger: logger,
		url:    url,
	}
}

func (handler *GetAllHandler) Handle(c *gin.Context) {
	resp, err := http.DefaultClient.Get(handler.url)
	if err != nil {
		handler.logger.Error(err)
		return
	}

	var courses []courseStruct
	err = json.NewDecoder(resp.Body).Decode(&courses)
	if err != nil {
		handler.logger.Error(err)
		return
	}

	outs := make([]outCourseStruct, len(courses))
	for i, course := range courses {
		outs[i] = outCourseStruct{
			ID:   course.ID,
			Src:  course.Body.Src,
			Text: course.Name,
			Data: course.Body.Data,
		}
	}

	c.JSON(http.StatusOK, outs)
}
