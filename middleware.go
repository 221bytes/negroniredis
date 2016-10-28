package negroniredis

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"gopkg.in/redis.v5"
)

const (
	Timeout  = 0
	Endpoint = 1
	Hybrid   = 2
)

type Middleware struct {
	http.ResponseWriter
	client *redis.Client
	key    string
	config Config
}

type Config struct {
	redisAddr           string
	redisPort           string
	redisPassword       string
	cacheExpirationTime time.Duration
	cacheStrategy       int
}

// default configuration
func DefaultConfig() Config {
	return Config{
		redisAddr:           "localhost",
		redisPort:           "6379",
		redisPassword:       "",
		cacheExpirationTime: time.Second * 2,
		cacheStrategy:       Hybrid,
	}
}

// Middleware is a struct that has a ServeHTTP method
func NewMiddleware(config Config) *Middleware {
	middlware := &Middleware{config: config}
	var buffer bytes.Buffer

	buffer.WriteString(config.redisAddr)
	buffer.WriteString(":")
	buffer.WriteString(config.redisPort)
	middlware.client = redis.NewClient(&redis.Options{
		Addr:     buffer.String(),
		Password: config.redisPassword,
		DB:       0,
	})
	pong, err := middlware.client.Ping().Result()
	fmt.Println(pong, err)
	return middlware
}

func (m *Middleware) Write(b []byte) (int, error) {
	err := m.client.Set(m.key, string(b), m.config.cacheExpirationTime).Err()
	if err != nil {
		panic(err)
	}
	return m.ResponseWriter.Write(b)
}

// The middleware handler
func (m *Middleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var buffer bytes.Buffer

	buffer.WriteString(req.Host)
	buffer.WriteString(":")
	buffer.WriteString(req.URL.RequestURI())
	m.key = buffer.String()
	ctxt := context.Background()
	client := m.client
	// scanner := client.Scan(0, "*", 100)

	cachedVal, err := client.Get(buffer.String()).Result()
	if err == redis.Nil {
		ctxt = context.WithValue(ctxt, "cache", nil)
	} else if err != nil {
		panic(err)
	} else {
		ctxt = context.WithValue(ctxt, "cache", cachedVal)
	}

	m.ResponseWriter = w
	if next != nil {
		next(m, req.WithContext(ctxt))
	}
}
