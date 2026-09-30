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

	// Adapters (concrete implementations of ports)
	serviceRepo := postgres.NewServiceRepo(pool)
	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := postgres.NewSessionRepo(pool)
	haircutRepo := postgres.NewHaircutRepo(pool)
	cashAdvanceRepo := postgres.NewCashAdvanceRepo(pool)
	hasher := bcryptadapter.NewHasher()

	// Use cases (application services)
	authSvc := app.NewAuthService(userRepo, hasher)
	sessionSvc := app.NewSessionService(sessionRepo, userRepo)
	haircutSvc := app.NewHaircutService(haircutRepo, serviceRepo)
	cashAdvanceSvc := app.NewCashAdvanceService(cashAdvanceRepo)

	// HTTP adapter
	handlers := &httpadapter.Handlers{
		Services:     serviceRepo,
		Users:        userRepo,
		Haircuts:     haircutSvc,
		CashAdvances: cashAdvanceSvc,
		Auth:         authSvc,
		Sessions:     sessionSvc,
	}
	router := httpadapter.NewRouter(handlers)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
