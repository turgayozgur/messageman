package main

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"log"
	"net"

	pb "github.com/turgayozgur/messageman/examples/job/workerapi/pb/v1/gen"
	"github.com/valyala/fasthttp"
)

type server struct {
	pb.UnimplementedWorkerServiceServer
}

func (server) Receive(context.Context, *pb.ReceiveRequest) (*empty.Empty, error) {
	log.Print("Email sent by using gRPC receiver.")
	return &empty.Empty{}, nil
}

func main() {
	// listen gRPC
	lis, err := net.Listen("tcp", ":83")
	if err != nil {
		panic(err)
	}
	gSrv := grpc.NewServer()
	pb.RegisterWorkerServiceServer(gSrv, &server{})
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
