package controllers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware() gin.HandlerFunc {

	//for i in {1..10}; do curl -X GET http://localhost:8080/ping & done

	rate := 3  // 5 requests per second
	ttl := 3 * time.Minute //timetolive , map is deleted after that
	
	type Visitor struct {
		requests int
		lastSeen time.Time
	}

	var mutex sync.Mutex
	visitors := make(map[string]*Visitor)

	go func() {
		for {
			time.Sleep(time.Minute)
			mutex.Lock()
			for VisitorIP, VisitorData := range visitors {
				if time.Since(VisitorData.lastSeen) > ttl {
					delete(visitors, VisitorIP)
				}
			}
			mutex.Unlock()
		}
	}()

	return func(c *gin.Context) {
		VisitorIP := c.ClientIP()

		mutex.Lock()
		VisitorData, exists := visitors[VisitorIP]
		if !exists {
			VisitorData = &Visitor{
				requests: 0, 
				lastSeen: time.Now(),
			}
			visitors[VisitorIP] = VisitorData
		}
		VisitorData.lastSeen = time.Now()

		if VisitorData.requests >= rate {
			mutex.Unlock() 
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":false,
				"message": "rate limit exceeded",
				"error_code":429,
				
			})
			return
		}
		VisitorData.requests++
		mutex.Unlock() 

		c.Next()

		go func() {
			time.Sleep(time.Second)
			mutex.Lock()
			VisitorData.requests--
			mutex.Unlock()
		}()
	}
}
