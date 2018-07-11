package handlers

import (
	"github.com/gin-gonic/gin"
)

func Refister(engine *gin.Engine) {
	engine.RouterGroup.Static()
}
