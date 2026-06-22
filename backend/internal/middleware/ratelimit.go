package middleware

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/queue"
	"github.com/gin-gonic/gin"
)

type RateLimitKey func(*gin.Context) string

func RateLimit(client *queue.Client, namespace string, limit int64, window time.Duration, key RateLimitKey) gin.HandlerFunc {
	return func(context *gin.Context) {
		identifier := key(context)
		allowed, remaining, retryAfter, err := client.Allow(
			context.Request.Context(),
			fmt.Sprintf("%s:%s", namespace, identifier),
			limit,
			window,
		)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "rate limiter is unavailable"})
			return
		}

		context.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		context.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		if !allowed {
			seconds := int(math.Ceil(retryAfter.Seconds()))
			if seconds < 1 {
				seconds = 1
			}
			context.Header("Retry-After", strconv.Itoa(seconds))
			context.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": seconds,
			})
			return
		}
		context.Next()
	}
}

func ClientIP(context *gin.Context) string {
	return context.ClientIP()
}

func AuthenticatedUser(context *gin.Context) string {
	if userID := context.GetString(ContextUserID); userID != "" {
		return userID
	}
	return context.ClientIP()
}
