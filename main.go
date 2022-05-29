package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "grpc-echo-client/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func callUnaryEcho(c pb.EchoClient, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := c.UnaryEcho(ctx, &pb.EchoRequest{Message: msg})
	if err != nil {
		fmt.Println("UnaryEcho: _, ", err)
	} else {
		fmt.Println("UnaryEcho: ", r.GetMessage())
	}
}

func sendMessage(stream pb.Echo_BidirectionalStreamingEchoClient, msg string) error {
	fmt.Printf("sending message %q\n", msg)
	return stream.Send(&pb.EchoRequest{Message: msg})
}

func recvMessage(stream pb.Echo_BidirectionalStreamingEchoClient, wantErrCode codes.Code) {
	res, err := stream.Recv()
	if status.Code(err) != wantErrCode {
		log.Fatalf("stream.Recv() = %v, %v; want _, status.Code(err)=%v", res, err, wantErrCode)
	}
	if err != nil {
		fmt.Printf("stream.Recv() returned expected error %v\n", err)
		return
	}
	fmt.Printf("received message %q\n", res.GetMessage())
}

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:5000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewEchoClient(conn)

	callUnaryEcho(c, "test")

	// Initiate the stream with a context that supports cancellation.
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := c.BidirectionalStreamingEcho(ctx)
	if err != nil {
		log.Fatalf("error creating stream: %v", err)
	}

	// Send some test messages.
	if err := sendMessage(stream, "hello"); err != nil {
		log.Fatalf("error sending on stream: %v", err)
	}
	if err := sendMessage(stream, "world"); err != nil {
		log.Fatalf("error sending on stream: %v", err)
	}

	// Ensure the RPC is working.
	recvMessage(stream, codes.OK)
	recvMessage(stream, codes.OK)

	fmt.Println("cancelling context")
	// cancel()

	// This Send may or may not return an error, depending on whether the
	// monitored context detects cancellation before the call is made.
	sendMessage(stream, "closed")

	// This Recv should never succeed.
	// recvMessage(stream, codes.Canceled)
}
