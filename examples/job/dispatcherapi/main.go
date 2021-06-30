package main

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"os"
	"time"

	pb "github.com/turgayozgur/messageman/examples/job/dispatcherapi/pb/v1/gen"
	"github.com/valyala/fasthttp"
)

var httpClient *http.Client
var emailQueueName = "send_email"
var gRPCCnn *grpc.ClientConn

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
			if string(ctx.QueryArgs().Peek("grpc")) == "true" {
				gRPCDispatchSendEmail()
			} else {
				dispatchSendEmail()
			}
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}
	log.Printf("Now, listening on: http://localhost:%s", "82")
	fasthttp.ListenAndServe(":82", m)
}

func dispatchSendEmail() {
	url := getEnv("MESSAGEMAN_URL", "http://messageman:8015")
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/queue?name=%s", url, emailQueueName), bytes.NewBuffer([]byte(`
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

func gRPCDispatchSendEmail() {
	if gRPCCnn == nil {
		url := getEnv("MESSAGEMAN_URL", "messageman:8020")
		if err := connGRPC(url); err != nil {
			log.Printf("%+v", err)
		}
	}

	c := pb.NewQueueServiceClient(gRPCCnn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"x-service-name": "dispatcherapi"}))
	if _, err := c.Queue(ctx, &pb.QueueRequest{
		Name: emailQueueName,
		Message: []byte(`
	{
		"email": "test@testmail.com"
	}`),
	}); err != nil {
		log.Printf("%+v", err)
	}

	log.Print("The send email job dispatched by using gRPC.")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func connGRPC(addr string) error {
	ctx := context.Background()
	var err error

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	log.Print(addr)
	cnn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return err
	}

	gRPCCnn = cnn

	return nil
}
