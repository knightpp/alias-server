package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func ZerologLogger(log zerolog.Logger) gin.HandlerFunc {
	log = log.With().Str("component", "middleware.gin").Logger()

	return func(c *gin.Context) {
		if !log.Trace().Enabled() {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		privateErrs := c.Errors.ByType(gin.ErrorTypePrivate)
		latency := time.Since(start)

		errs := make([]error, len(privateErrs))
		for i, err := range privateErrs {
			errs[i] = err
		}

		log.Trace().
			Errs("private_errors", errs).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("body.size", c.Writer.Size()).
			Msg("handling request")
	}
}
