package main

import (
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	midtrans2 "github.com/midtrans/midtrans-go"
	"log"
	"net/http"
	"payment-service-iae/config"
	"payment-service-iae/graph"
	"payment-service-iae/midtrans"
)

func main() {

	cfg := config.Load()
	port := cfg.Port

	midtransClient := midtrans.NewClient(
		cfg.MidtransServerKey,
		midtrans2.Sandbox,
	)

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "user", "user-123")
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver(midtransClient)}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", authMiddleware(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
