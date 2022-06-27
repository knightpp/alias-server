package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knightpp/alias-server/internal/storage"
)

const UserIDKey = "user-id"

func Authorized(db storage.PlayerDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("auth")

		var err error
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err = db.GetPlayer(ctx, auth)
		}()
		if err != nil {
			c.String(http.StatusForbidden, "you must authorize to access this resource")
			c.Abort()
			return
		}

		c.Set(UserIDKey, auth)

		c.Next()
	}
}
