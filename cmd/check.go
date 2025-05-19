package main

import (
	"context"
	"fmt"
	"github.com/ecociel/httprouterext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

func main() {

	//ns := "customer"
	//obj := "foo"
	//rel := "customer.get"
	//userId := "723B7781-7C28-4F63-821B-9C652DBA482C"
	ns := os.Args[1]
	obj := os.Args[2]
	rel := os.Args[3]
	userId := os.Args[4]

	hostport := "localhost:50052"

	conn, err := grpc.NewClient(hostport, grpc.
		WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect check-service at %q: %v", hostport, err)
	}

	c := httprouterext.New(conn)

	principal, ok, err := c.CheckWithTimestamp(context.Background(), httprouterext.Namespace(ns), httprouterext.Obj(obj), httprouterext.Permission(rel), httprouterext.UserId(userId), httprouterext.Timestamp("1:0000000000000"))
	if err != nil {
		log.Fatalf("check: %v", err)
	}
	fmt.Printf("Result: %s %v", principal, ok)

}
