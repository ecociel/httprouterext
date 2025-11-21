# Go Client for NIO Authorization

[](https://www.google.com/search?q=https://goreportcard.com/report/github.com/ecociel/httprouterext)

A Go client library for the **NIO Authorization Service**, a high-performance, relationship-based authorization system. This library provides a clean gRPC client to interact with the service and includes middleware for easy integration with the `julienschmidt/httprouter` framework.

-----

## Features

  - ✅ **Core gRPC Client:** A simple and robust client for all `CheckService` RPCs (`Check`, `List`, `Read`, `Write`).
  - 🔌 **`httprouter` Middleware:** A powerful wrapper (`Wrap`) that automates authorization checks for your HTTP routes.
  - 🧑‍💻 **Request-Scoped User Object:** Access a `User` object within your handlers to perform fine-grained permission checks.
  - ⏱️ **Observability Hooks:** Provide custom functions to meter the latency and results of authorization checks.
  - 📝 **Custom Error Handling:** Map authorization errors to specific HTTP responses.

-----

## Installation

```sh
go get github.com/ecociel/httprouterext
```

-----

## Getting Started

Here is a complete example of how to protect an `httprouter` route using this library.

### 1\. Define your Resource

First, create a struct that represents a resource in your application and implement the `Resource` interface. The `Requires` method defines the permission needed to access it.

```go
// in your application, e.g., in a file named article.go

package main

import (
    "fmt"
    "net/http"
    "github.com/ecociel/httprouterext"
    "github.com/julienschmidt/httprouter"
)

// ArticleResource represents a single article, identified by its ID.
type ArticleResource struct {
    ID string
}

// Requires defines the permission needed to access an article.
// For a GET request, it requires 'article.get' permission.
func (a *ArticleResource) Requires(principalOrToken string, method string) (
    ns httprouterext.Namespace, 
    obj httprouterext.Obj, 
    permission httprouterext.Permission,
) {
    ns = "articles"
    obj = httprouterext.Obj(fmt.Sprintf("article:%s", a.ID))
    
    switch method {
    case http.MethodGet:
        permission = "article.get"
    case http.MethodPut:
        permission = "article.update"
    default:
        permission = httprouterext.Impossible // Deny by default
    }
    return
}
```

### 2\. Set Up the Client and Middleware

In your `main.go`, set up the gRPC client, create the router, and wrap your handler.

```go
// main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ecociel/httprouterext"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// This function extracts the resource from the incoming request.
func articleExtractor(r *http.Request, p httprouter.Params) (httprouterext.Resource, error) {
	articleID := p.ByName("id")
	if articleID == "" {
		return nil, fmt.Errorf("article ID is missing")
	}
	return &ArticleResource{ID: articleID}, nil
}

// This is your actual application logic.
func getArticleHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params, resource httprouterext.Resource, user httprouterext.User) error {
	// The initial 'article.get' check has already passed.
	// Now, you can perform more specific checks if needed.
	
    // Example: Check if the user can also comment on this article.
	canComment, err := user.HasPermission("article.comment")
	if err != nil {
		return fmt.Errorf("failed to check comment permission: %w", err)
	}

	article := resource.(*ArticleResource)
	response := fmt.Sprintf("Hello, %s! You are viewing article %s.", user.Principal(), article.ID)
	
    if canComment {
		response += " You are also allowed to comment."
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
	return nil
}


func main() {
	// 1. Connect to the NIO Authorization gRPC service
	authzHost := "localhost:50052"
	conn, err := grpc.Dial(authzHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to authz service: %v", err)
	}
	defer conn.Close()

	// 2. Create a new NIO client
	nioClient := httprouterext.New(conn)

	// 3. Set up httprouter
	router := httprouter.New()
    
    // This is a placeholder for your real authentication logic.
    // The wrapper expects a function that can extract a user token from the request.
    tokenExtractor := func(r *http.Request) (string, error) {
        // In a real app, you would get this from a cookie or "Authorization" header.
        sessionCookie, err := r.Cookie("session")
        if err != nil {
            return "", err
        }
        return sessionCookie.Value, nil
    }

	// 4. Wrap your handler with the authorization middleware
	router.GET("/articles/:id", httprouterext.Wrap(nioClient, tokenExtractor, articleExtractor, getArticleHandler))

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

```

-----

## Working with Protobuf

The gRPC service definitions are located in the `proto/` directory. If you modify the `.proto` files, you will need to regenerate the Go code.

We recommend using `go:generate` for this. Create a file named `proto/generate.go` with the following content:

```go
// proto/generate.go
package proto

//go:generate protoc --proto_path=. --go_out=. --go-grpc_out=. iam.proto
```

Now, you can regenerate the code at any time by running:

```sh
go generate ./...
```

-----

## Tsting

admin@local.local|226ab259-cf37-4cf9-877f-71ee0a795243
anna@local.local|a63c2dff-6cdb-4625-b3c7-250bfdc44bb8
bob@local.local|7fcc7383-2999-491a-ae7d-68d3497936f4

9  go run cmd/list.go project project.get bbf15352-f4d8-4c83-8043-b1449ed77ae2
  540  go run cmd/write.go project p44 owner 89433e8d-adbf-45e8-a6b6-1b59cfe96831
  541  go run cmd/list.go project project.get 89433e8d-adbf-45e8-a6b6-1b59cfe96831
  542  go run cmd/write.go project p43 owner 89433e8d-adbf-45e8-a6b6-1b59cfe96831
  543  go run cmd/write.go project p49 owner 89433e8d-adbf-45e8-a6b6-1b59cfe96831
  544  go run cmd/list.go project project.get 89433e8d-adbf-45e8-a6b6-1b59cfe96831
  545  go run cmd/write.go serviceaccount deploy parent 'root:root#...'
  546  go run cmd/write.go serviceaccount deploy parent 'root:root#...'
  547  sqlite3 test/check.sqlite 'select * from tuples_root,
  548  sqlite3 test/check.sqlite 'select * from tuples_root'
  549  sqlite3 test/check.sqlite 'select * from tuples_serviceaccount'
  550  go run cmd/list.go serviceaccount serviceaccount.get bbf15352-f4d8-4c83-8043-b1449ed77ae2


## License

This project is licensed under the **Apache 2.0 License**. See the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.
