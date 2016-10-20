package negroniredis

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"gopkg.in/redis.v5"
)

type Middleware struct {
	client *redis.Client
	http.ResponseWriter
	key string
}

// Middleware is a struct that has a ServeHTTP method
func NewMiddleware() *Middleware {
	middlware := &Middleware{}
	middlware.client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	pong, err := middlware.client.Ping().Result()
	fmt.Println(pong, err)
	return middlware
}

func (m *Middleware) Write(b []byte) (int, error) {
	err := m.client.Set(m.key, string(b)+"cached", 0).Err()
	if err != nil {
		panic(err)
	}
	return m.ResponseWriter.Write(b)
}

// The middleware handler
func (m *Middleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Printf("Req: %s%s\n", req.Host, req.URL.Path)
	var buffer bytes.Buffer

	buffer.WriteString(req.Host)
	buffer.WriteString(":")
	buffer.WriteString(req.RequestURI)
	m.key = buffer.String()
	ctxt := context.Background()
	client := m.client

	val2, err := client.Get(buffer.String()).Result()
	if err == redis.Nil {
		ctxt = context.WithValue(ctxt, "cache", nil)
	} else if err != nil {
		panic(err)
	} else {
		ctxt = context.WithValue(ctxt, "cache", val2)
	}

	m.ResponseWriter = w
	if next != nil {
		next(m, req.WithContext(ctxt))

	}
}
