package negroniredis

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gopkg.in/redis.v5"
)

const (
	ContextKey = "NEGRONISREDISCACHE"
)

type Middleware struct {
	http.ResponseWriter
	client *redis.Client
	config Config
}

type Config struct {
	redisAddr           string
	redisPort           string
	redisPassword       string
	cacheExpirationTime time.Duration
	prefix              string
}

var middleware *Middleware
var once sync.Once

// default configuration
func DefaultConfig() Config {
	return Config{
		redisAddr:           "localhost",
		redisPort:           "6379",
		redisPassword:       "",
		cacheExpirationTime: time.Second * 2,
		prefix:              "cache",
	}
}

// Middleware is a struct that has a ServeHTTP method
func NewMiddleware(config Config) *Middleware {
	once.Do(func() {

		middleware = &Middleware{config: config}
		var buffer bytes.Buffer

		buffer.WriteString(config.redisAddr)
		buffer.WriteString(":")
		buffer.WriteString(config.redisPort)
		middleware.client = redis.NewClient(&redis.Options{
			Addr:     buffer.String(),
			Password: config.redisPassword,
			DB:       0,
		})
		pong, err := middleware.client.Ping().Result()
		fmt.Println(pong, err)
	})
	return middleware
}

type Writer struct {
	http.ResponseWriter
	key        string
	reqContext context.Context
}

func (w *Writer) Write(b []byte) (int, error) {
	// if request is already from the cache we shouldn't cache it again

	if cache := w.reqContext.Value(ContextKey); cache != nil {
		return w.ResponseWriter.Write(b)
	}
	// we cache new data
	err := middleware.client.Set(w.key, string(b), middleware.config.cacheExpirationTime).Err()
	if err != nil {
		panic(err)
	}

	return w.ResponseWriter.Write(b)
}

// The middleware handler
func (m *Middleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var buffer bytes.Buffer

	buffer.WriteString(m.config.prefix)
	buffer.WriteString(":")
	buffer.WriteString(req.Host)
	buffer.WriteString(req.URL.RequestURI())
	ctxt := context.Background()
	client := m.client

	// scanner := client.Scan(0, "*", 100)

	cachedVal, err := client.Get(buffer.String()).Result()
	if err == redis.Nil {
		ctxt = context.WithValue(ctxt, ContextKey, nil)
	} else if err != nil {
		panic(err)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		ctxt = context.WithValue(ctxt, ContextKey, cachedVal)
	}

	writer := &Writer{ResponseWriter: w, key: buffer.String(), reqContext: ctxt}
	if next != nil {
		next(writer, req.WithContext(ctxt))
	}
}
