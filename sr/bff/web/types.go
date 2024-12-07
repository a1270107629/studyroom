package web

import (
	"github.com/a1270107629/studyroom/sr/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type Result = ginx.Result

type Page struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type handler interface {
	RegisterRoutes(s *gin.Engine)
}
