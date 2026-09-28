package main

import (
	"context"
	"log"
	"net/http"

	httpadapter "barbershop/internal/adapters/http"
	"barbershop/internal/adapters/postgres"
)

func main() {
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// The ONLY place that wires concrete types together.
	serviceRepo := postgres.NewServiceRepo(pool)
	handlers := &httpadapter.Handlers{Services: serviceRepo}
	router := httpadapter.NewRouter(handlers)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
