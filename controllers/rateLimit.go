package controllers

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware() gin.HandlerFunc {
    apiRate := 3
	ttl := 3 * time.Minute //time-to-live,map with that value are deleted after that

	type Visitor struct {
		requests int
		lastSeen time.Time
	}

	var (
		mutex    sync.Mutex
		visitors = make(map[string]*Visitor)
	)

	return func(c *gin.Context) {
		visitorIP := c.ClientIP()
		mutex.Lock()
		visitorData, exists := visitors[visitorIP]
		if !exists {
			visitorData = &Visitor{
				requests: 0,
				lastSeen: time.Now(),
			}
			visitors[visitorIP] = visitorData
		}
		visitorData.lastSeen = time.Now()
		if visitorData.requests >= apiRate {
			mutex.Unlock()
			message := fmt.Sprintf("rate limit exceeded for IP: %v", visitorIP)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":     false,
				"message":    message,
				"error_code": http.StatusTooManyRequests,
			})
			return
		}
		visitorData.requests++
		mutex.Unlock()

		c.Next()

		
	//decrement the requestcount after each second based on the apirate
	go func() {
		time.Sleep(time.Second)
		mutex.Lock()
		if visitorData.requests != 0{
			visitorData.requests-= apiRate
		}
		mutex.Unlock()
	}()

	//deleting the map with the specific ip as key if the last seen is greater than ttl -> ip not active more than the ttl time
	go func() {
		for {
			time.Sleep(time.Minute)
			mutex.Lock()
			for ip, visitor := range visitors {
				if time.Since(visitor.lastSeen) > ttl {
					delete(visitors, ip)
				}
			}
			mutex.Unlock()
		}
	}()
	}
}




//getting client ip
//rate,ttl,map,visitor struct
//create ,map[ip]visitor, 
//visitor - requestcount,lasttime
//after each request increment requestcount
//go routines- 
  //one for decreasing the apirate count of the ip after each sec
  //delete the ip address from the map if the ttl is greater than the time.since(lastseen) to the ttl
  //active time is less than 3 minutes delete the ip