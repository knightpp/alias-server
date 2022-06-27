package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func ZerologLogger(log zerolog.Logger) gin.HandlerFunc {
	log = log.With().Str("component", "middleware.gin").Logger()

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		var event *zerolog.Event
		if c.Writer.Status() != 200 {
			event = log.Error()
		} else {
			event = log.Trace()
		}
		if !event.Enabled() {
			return
		}

		privateErrs := c.Errors.ByType(gin.ErrorTypePrivate)
		latency := time.Since(start)
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		errs := make([]error, len(privateErrs))
		for i, err := range privateErrs {
			errs[i] = err
		}

		event.
			Int("status_code", c.Writer.Status()).
			Str("status_text", http.StatusText(c.Writer.Status())).
			Errs("private_errors", errs).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("user_agent", c.Request.UserAgent()).
			Int("body.size", c.Writer.Size()).
			Msg("handled request")
	}
}
