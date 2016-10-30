package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/221bytes/negroniredis"
	"github.com/221bytes/negroniredis/cachegroup"
	"github.com/urfave/negroni"
)

type exampleStruct struct {
	I int  `json:"id"`
	A bool `json:"test"`
}

func main() {
	mux := http.NewServeMux()
	cg0 := cachegroup.CreateCacheGroup("/", "/test", "/toto")
	cgm := cachegroup.NewCacheGroupManager()
	cgm.AddCacheGroup(cg0)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		if cache := req.Context().Value(negroniredis.ContextKey); cache != nil {
			fmt.Fprint(w, cache)
			return
		}
		time.Sleep(time.Millisecond * 500)

		foo := exampleStruct{A: false, I: 32}
		if err := json.NewEncoder(w).Encode(&foo); err != nil {
			panic(err)
		}
	})

	n := negroni.Classic()

	redisCacheConfig := negroniredis.DefaultConfig()
	redisCacheConfig.CGM = cgm
	n.Use(negroniredis.NewMiddleware(redisCacheConfig))

	n.UseHandler(mux)

	n.Run(":3000")
}
