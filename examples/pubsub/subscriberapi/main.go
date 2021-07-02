package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"log"
	"net"

	pb "github.com/turgayozgur/messageman/examples/pubsub/subscriberapi/pb/v1/gen"
	"github.com/valyala/fasthttp"
)

type server struct {
	pb.UnimplementedHandlerServiceServer
}

func (server) Handle(ctx context.Context, in *pb.HandleRequest) (*empty.Empty, error) {
	log.Printf("%s event handled by using gRPC handler.", in.Name)
	return &empty.Empty{}, fmt.Errorf("error")
}

func main() {
	// listen gRPC
	lis, err := net.Listen("tcp", ":83")
	if err != nil {
		panic(err)
	}
	gSrv := grpc.NewServer()
	pb.RegisterHandlerServiceServer(gSrv, &server{})
	go func() {
		log.Printf("Now, gRPC listening on: http://localhost:83")
		if err := gSrv.Serve(lis); err != nil {
			log.Printf("%v", err)
		}
	}()

	// listen REST
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
	log.Printf("Now, listening on: http://localhost:%s", "81")
	fasthttp.ListenAndServe(":81", m)
}
