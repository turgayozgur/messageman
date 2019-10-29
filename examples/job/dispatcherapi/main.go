package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/valyala/fasthttp"
)

var httpClient *http.Client
var emailQueueName = "email"

func main() {
	httpClient = &http.Client{
		Timeout: time.Second * 60,
	}

	m := func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if rc := recover(); rc != nil {
				log.Printf("Panic during request. %+v", rc)
			}
		}()
		switch string(ctx.Path()) {
		case "/readiness":
			ctx.Response.SetStatusCode(fasthttp.StatusOK)
		case "/api/order/create":
			//time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
			log.Print("Order created.")
			dispatchSendEmail()
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}
	log.Printf("Now listening on: http://localhost:%s", "82")
	fasthttp.ListenAndServe(":82", m)
}

func dispatchSendEmail() {
	url := getEnv("MESSAGEMAN_URL", "http://messageman:8015")
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/queue?name=%s", url, emailQueueName), bytes.NewBuffer([]byte(`
	{
		"email": "test@testmail.com"
	}`)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-service-name", "dispatcherapi") // good to know for gateway implementation of messageman.
	response, err := httpClient.Do(req)
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	if response.StatusCode != 200 {
		log.Printf("%+v", response.StatusCode)
		return
	}
	log.Print("The send email job dispatched.")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
