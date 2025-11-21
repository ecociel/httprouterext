package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ecociel/httprouterext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ns := os.Args[1]
	obj := os.Args[2]
	rel := os.Args[3]
	user := os.Args[4]

	hostport := "localhost:50052"

	conn, err := grpc.NewClient(hostport, grpc.
		WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect check-service at %q: %v", hostport, err)
	}

	c := httprouterext.New(conn)

	if strings.Contains(user, "#") {
		a := strings.SplitN(user, ":", 2)
		println(a)
		b := strings.SplitN(a[1], "#", 2)
		println(b)

		userSet := httprouterext.UserSet{
			Ns:  httprouterext.Namespace(a[0]),
			Obj: httprouterext.Obj(b[0]),
			Rel: httprouterext.Permission(b[1]),
		}
		fmt.Printf("%v\n", userSet)
		err = c.AddOneUserSet(context.Background(),
			httprouterext.Namespace(ns), httprouterext.Obj(obj), httprouterext.Permission(rel), userSet)
	} else {

		err = c.AddOneUserId(context.Background(),
			httprouterext.Namespace(ns), httprouterext.Obj(obj), httprouterext.Permission(rel), httprouterext.UserId(user))
	}
	if err != nil {
		log.Fatalf("add-one: %v", err)
	}

}
