package main

import (
	"log"

	"github.com/valyala/fasthttp"
)

func main() {
	m := func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if rc := recover(); rc != nil {
				log.Printf("Panic during request. %+v", rc)
			}
		}()
		switch string(ctx.Path()) {
		case "/readiness":
			ctx.Response.SetStatusCode(fasthttp.StatusOK)
		case "/api/email/send":
			//time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
			log.Print("Email sent.")
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}
	log.Printf("Now listening on: http://localhost:%s", "81")
	fasthttp.ListenAndServe(":81", m)
}
