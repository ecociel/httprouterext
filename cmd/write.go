package main

import (
	"context"
	"log"
	"os"

	"github.com/ecociel/httprouterext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ns := os.Args[1]
	obj := os.Args[2]
	rel := os.Args[3]
	userId := os.Args[4]

	hostport := "localhost:50051"

	conn, err := grpc.NewClient(hostport, grpc.
		WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect check-service at %q: %v", hostport, err)
	}

	c := httprouterext.New(conn)

	err = c.AddOneUserId(context.Background(),
		httprouterext.Namespace(ns), httprouterext.Obj(obj), httprouterext.Permission(rel), httprouterext.UserId(userId))
	if err != nil {
		log.Fatalf("add-one: %v", err)
	}

}
