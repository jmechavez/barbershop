package main

import (
	"context"
	"log"
	"net/http"

	bcryptadapter "barbershop/internal/adapters/bcrypt"
	httpadapter "barbershop/internal/adapters/http"
	"barbershop/internal/adapters/postgres"
	"barbershop/internal/app"
)

func main() {
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Adapters (the concrete implementations of our ports).
	serviceRepo := postgres.NewServiceRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := postgres.NewSessionRepo(pool)
	hasher := bcryptadapter.NewHasher()

	// Use cases (the app layer).
	authSvc := app.NewAuthService(userRepo, hasher)
	sessionSvc := app.NewSessionService(sessionRepo, userRepo)

	// HTTP adapter.
	handlers := &httpadapter.Handlers{
		Services: serviceRepo,
		Users:    userRepo,
		Auth:     authSvc,
		Sessions: sessionSvc,
	}
	router := httpadapter.NewRouter(handlers)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
