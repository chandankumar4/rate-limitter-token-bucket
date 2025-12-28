package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type tokenBucket struct {
	mu             sync.Mutex
	tokens         float64
	maxToken       float64
	refillRate     float64
	lastRefillTime time.Time
}

func NewTokenBucket(maxToken, refillRate float64) *tokenBucket {
	return &tokenBucket{
		tokens:         maxToken,
		maxToken:       maxToken,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

func (tb *tokenBucket) refill() {
	now := time.Now()
	durantion := now.Sub(tb.lastRefillTime)
	tokenToAdd := tb.refillRate * durantion.Seconds()
	tb.tokens = math.Min(tb.tokens+tokenToAdd, tb.maxToken)
	tb.lastRefillTime = now
}

func (tb *tokenBucket) Request(tokens float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	if tokens <= tb.tokens {
		tb.tokens -= tokens
		return true
	}

	return false
}

type userTokenBucket struct {
	userTokenBuckets map[string]*tokenBucket
	mutex            sync.Mutex
}

func NewUserTokenBucketManager() *userTokenBucket {
	return &userTokenBucket{
		userTokenBuckets: make(map[string]*tokenBucket),
	}
}

func (utb *userTokenBucket) getUserTokenBucket(ip string) *tokenBucket {
	utb.mutex.Lock()
	defer utb.mutex.Unlock()

	if bucket, ok := utb.userTokenBuckets[ip]; ok {
		return bucket
	}
	fmt.Println("new bucket created---------------")

	bucket := NewTokenBucket(5, 1)
	utb.userTokenBuckets[ip] = bucket

	return bucket
}

func (utb *userTokenBucket) requstFromUser(ip string) bool {
	userTokenBucket := utb.getUserTokenBucket(ip)
	utb.mutex.Lock()
	defer utb.mutex.Unlock()

	userTokenBucket.refill()
	if userTokenBucket.tokens >= 1 {
		userTokenBucket.tokens = userTokenBucket.tokens - 1
		return true
	}

	return false
}

var ServiceATokenBucket = NewTokenBucket(10, 3)
var ServiceBTokenBucket = NewTokenBucket(20, 2)

func main() {
	userTokenBucket := NewUserTokenBucketManager()

	for range 1000 {
		go func() {
			if ServiceATokenBucket.Request(1) {
				fmt.Println("Service A Request Accepted")
			} else {
				fmt.Println("Service A Request Denied")
			}
		}()

		go func() {
			if ServiceBTokenBucket.Request(1) {
				fmt.Println("Service B Request Accepted")
			} else {
				fmt.Println("Service B Request Denied")
			}
		}()
		go func() {
			ip := "192.168.1.1"
			if userTokenBucket.requstFromUser(ip) {
				fmt.Println("User Request Accepted")
			} else {
				fmt.Println("User Request Denied")
			}
		}()

		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(5 * time.Second)
}
