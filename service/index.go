package service

import (
	"ginchat/models"

	"github.com/gin-gonic/gin"
)

func Chat(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}
