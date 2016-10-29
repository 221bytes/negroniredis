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
	// ContextKey is for finding cached data in the Context
	ContextKey = "NEGRONISREDISCACHE"

	// httpMethod is for storing the method of the request in the reqWriter
	httpMethod = "HTTP_METHOD"
)

var middleware *RedisCache
var once sync.Once

// reqWriter is for intercepting the write of the http.ResponseWriter of the request and
// save the answer to the cache
// each request will have a different reqWriter
type reqWriter struct {
	http.ResponseWriter
	key        string
	reqContext context.Context
}

func (w *reqWriter) Write(b []byte) (int, error) {

	// we cache only get data
	if reqMethod := w.reqContext.Value(httpMethod); reqMethod != http.MethodGet {
		return w.ResponseWriter.Write(b)
	}

	// if request is already from the cache we shouldn't cache it again
	if cache := w.reqContext.Value(ContextKey); cache != nil {
		return w.ResponseWriter.Write(b)
	}
	// we cache new data
	err := middleware.RedisClient.Set(w.key, string(b), middleware.Config.CacheExpirationTime).Err()
	if err != nil {
		panic(err)
	}

	return w.ResponseWriter.Write(b)
}

// RedisCache is the middleware for negroniredis
type RedisCache struct {
	RedisClient *redis.Client
	Config      Config
}

// Config is all of the required fields needed by the cache
// RedisAddr, RedisPort and RedisPassword are the parameters
// for the redis library 	"gopkg.in/redis.v5"
// cacheExpirationTime is for how long each data will stay in the cache
// prefix is the prefix of every keys in the cache
type Config struct {
	RedisAddr           string
	RedisPort           string
	RedisPassword       string
	CacheExpirationTime time.Duration
	Prefix              string
}

// DefaultConfig is basic configuration for a RedisCache
func DefaultConfig() Config {
	return Config{
		RedisAddr:           "localhost",
		RedisPort:           "6379",
		RedisPassword:       "",
		CacheExpirationTime: time.Second * 2,
		Prefix:              "cache",
	}
}

// NewMiddleware return a RedisCache based on a configuration
func NewMiddleware(config Config) *RedisCache {
	once.Do(func() {

		middleware = &RedisCache{Config: config}
		var buffer bytes.Buffer

		buffer.WriteString(config.RedisAddr)
		buffer.WriteString(":")
		buffer.WriteString(config.RedisPort)
		middleware.RedisClient = redis.NewClient(&redis.Options{
			Addr:     buffer.String(),
			Password: config.RedisPassword,
			DB:       0,
		})
		pong, err := middleware.RedisClient.Ping().Result()
		fmt.Println(pong, err)
	})
	return middleware
}

func handleGet(client *redis.Client, key string) context.Context {
	ctxt := context.Background()
	cachedVal, err := client.Get(key).Result()
	if err == redis.Nil {
		ctxt = context.WithValue(ctxt, ContextKey, nil)
	} else if err != nil {
		panic(err)
	} else {
		ctxt = context.WithValue(ctxt, ContextKey, cachedVal)
	}
	return ctxt
}

func handleModif(client *redis.Client, key string) context.Context {
	ctxt := context.Background()
	fmt.Println(key)
	err := client.Del(key).Err()
	if err == redis.Nil {
	} else if err != nil {
		panic(err)
	} else {
	}
	return ctxt
}

// Basic negroni middleware function
func (m *RedisCache) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var buffer bytes.Buffer

	buffer.WriteString(m.Config.Prefix)
	buffer.WriteString(":")
	buffer.WriteString(req.Host)
	buffer.WriteString(req.URL.RequestURI())
	key := buffer.String()
	client := m.RedisClient

	var ctxt context.Context
	if req.Method == http.MethodGet {
		ctxt = handleGet(client, key)
	} else {
		ctxt = handleModif(client, key)
	}
	ctxt = context.WithValue(ctxt, httpMethod, req.Method)

	writer := &reqWriter{ResponseWriter: w, key: key, reqContext: ctxt}
	if next != nil {
		next(writer, req.WithContext(ctxt))
	}
}
