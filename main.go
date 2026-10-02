package main

import (
	"context"
	"log"
	"net/http"
	"os"

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
	salarySvc := app.NewSalaryService(haircutRepo, cashAdvanceRepo, userRepo)

	// HTTP adapter
	handlers := &httpadapter.Handlers{
		Services:     serviceRepo,
		Users:        userRepo,
		Haircuts:     haircutSvc,
		CashAdvances: cashAdvanceSvc,
		Auth:         authSvc,
		Sessions:     sessionSvc,
		Salary:       salarySvc,
	}
	router := httpadapter.NewRouter(handlers)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
