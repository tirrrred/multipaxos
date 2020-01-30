// +build !solution

// Leave an empty line above this comment.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"google.golang.org/grpc"

	pb "github.com/dat520-2020/assignments/lab2/grpc/proto"
)

var (
	help = flag.Bool(
		"help",
		false,
		"Show usage help",
	)
	endpoint = flag.String(
		"endpoint",
		"localhost:12111",
		"Endpoint on which server runs or to which client connects",
	)
)

type kvData struct {
	kv map[string]string
}

func main() {
	ctx := context.Background()
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Grpc connection failed:", err)
	}

	defer conn.Close()
	client := pb.NewKeyValueServiceClient(conn)

	InsResp1, err := client.Insert(ctx, &pb.InsertRequest{"Key1", "Value1"})
	InsResp2, err := client.Insert(ctx, &pb.InsertRequest{"Key2", "Value2"})
	client.Insert(ctx, &pb.InsertRequest{"RandomKey", "EvenMoreRandomValue"})
	client.Insert(ctx, &pb.InsertRequest{"NewKeyAfterMutex", "NewValueAfterMutex"})

	fmt.Printf("\nGot response: %v (Type: %T)", InsResp1, InsResp1)
	fmt.Printf("\nGot response: %v (Type: %T)", InsResp2, InsResp2)

	LooResp, err := client.Lookup(ctx, &pb.LookupRequest{"Key1"})

	fmt.Printf("\nValue from lookup: %v\n", LooResp)

	KeyResp, err := client.Keys(ctx, &pb.KeysRequest{})

	fmt.Printf("\nKeys from keys: %v\n", KeyResp)

}
