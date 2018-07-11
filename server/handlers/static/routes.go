package static

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
)

// Register adds static routes such as not found
func Register(engine *gin.Engine) {
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, base.ErrorView{
			Message: "Not found",
		})
	})
}
