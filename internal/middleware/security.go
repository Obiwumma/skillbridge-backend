// Package middleware handles HTTP request preprocessing.
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders injects standard secure HTTP response headers to prevent frame attacks and hijacking.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}

// CORS maps Cross-Origin Resource Sharing approvals and instantly satisfies preflight OPTIONS requests.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger outputs metadata and latency measurements for unsuccessful requests.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if status >= 400 {
			log.Printf("[API-FAILURE] | Status: %d | Latency: %13v | ClientIP: %s | Route: %s %s %s",
				status, latency, c.ClientIP(), c.Request.Method, path, query,
			)
		}
	}
}
