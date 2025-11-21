package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ecociel/httprouterext"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ArticleResource represents a single article, identified by its ID.
type ArticleResource struct {
	ID string
}

func (r *ArticleResource) Link() string {
	return fmt.Sprintf("/articles/%s", r.ID)
}

func (r *ArticleResource) Requires(principalOrToken string, method string) (httprouterext.Namespace, httprouterext.Obj, httprouterext.Permission) {
	var permission httprouterext.Permission
	switch method {
	case http.MethodHead:
	case http.MethodGet:
		permission = "article.get"
	case http.MethodPost:
		permission = "article.update"
	default:
		permission = httprouterext.Impossible
	}

	return httprouterext.Namespace("article"), httprouterext.Obj(r.ID), permission
}

func ExtractArticleResource(r *http.Request, p httprouter.Params) (httprouterext.Resource, error) {
	id := p.ByName("id")
	if id == "" {
		panic("wrong router configuration")
	}
	return &ArticleResource{
		ID: id,
	}, nil
}

var RouteArticleResource = ArticleResource{
	ID: ":id",
}

func getArticle(w http.ResponseWriter, r *http.Request, p httprouter.Params, resource httprouterext.Resource, user httprouterext.User) error {
	articleResource := resource.(*ArticleResource)
	fmt.Fprintf(w, "Article id=%s", articleResource.ID)
	return nil
}

func main() {
	// 1. Connect to the NIO Authorization gRPC service
	checkHostPort := "localhost:50051"

	conn, err := grpc.NewClient(checkHostPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect check-service at %q: %v", checkHostPort, err)
	}

	nioClient := httprouterext.New(conn)

	router := httprouter.New()

	router.GET(RouteArticleResource.Link(), httprouterext.Wrap(nioClient, ExtractArticleResource, getArticle))

	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe("127.0.0.1:8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
