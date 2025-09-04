package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
	rate    int           // 每分钟允许的请求数
	window  time.Duration // 时间窗口
}

// ClientInfo 客户端信息
type ClientInfo struct {
	requests  int
	lastReset time.Time
}

var limiter *RateLimiter

func init() {
	limiter = &RateLimiter{
		clients: make(map[string]*ClientInfo),
		rate:    60, // 每分钟60个请求
		window:  time.Minute,
	}

	// 定期清理过期的客户端记录
	go limiter.cleanup()
}

// RateLimit 速率限制中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow 检查是否允许请求
func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientIP]

	if !exists {
		// 新客户端
		rl.clients[clientIP] = &ClientInfo{
			requests:  1,
			lastReset: now,
		}
		return true
	}

	// 检查是否需要重置计数器
	if now.Sub(client.lastReset) >= rl.window {
		client.requests = 1
		client.lastReset = now
		return true
	}

	// 检查是否超过速率限制
	if client.requests >= rl.rate {
		return false
	}

	client.requests++
	return true
}

// cleanup 定期清理过期的客户端记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, client := range rl.clients {
				if now.Sub(client.lastReset) > 2*rl.window {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}