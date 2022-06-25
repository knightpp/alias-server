package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knightpp/alias-server/internal/data"
	"github.com/knightpp/alias-server/internal/game"
)

func Authorized(game *game.Game) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("auth")

		if !game.IsPlayerExists(data.PlayerID(auth)) {
			c.String(http.StatusForbidden, "you must authorize to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}
