package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

type fakeSessions struct {
	byToken map[string]fakeSession
}

type fakeSession struct {
	userID    int
	expiresAt time.Time
}

func (f *fakeSessions) Create(ctx context.Context, token string, userID int, expiresAt time.Time) error {
	f.byToken[token] = fakeSession{userID: userID, expiresAt: expiresAt}
	return nil
}

func (f *fakeSessions) FindUserID(ctx context.Context, token string) (int, error) {
	s, ok := f.byToken[token]
	if !ok || time.Now().After(s.expiresAt) {
		return 0, ports.ErrSessionNotFound
	}
	return s.userID, nil
}

func (f *fakeSessions) Delete(ctx context.Context, token string) error {
	delete(f.byToken, token)
	return nil
}

func TestSessionStartAndResolve(t *testing.T) {
	users := &fakeUsers{byEmail: map[string]domain.User{
		"juan@shop.ph": {ID: 42, Email: "juan@shop.ph", Role: domain.RoleBarber},
	}}
	sessions := &fakeSessions{byToken: map[string]fakeSession{}}

	svc := NewSessionService(sessions, users)

	token, err := svc.Start(context.Background(), 42)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if token == "" {
		t.Fatal("Start returned empty token")
	}

	u, err := svc.UserForToken(context.Background(), token)
	if err != nil {
		t.Fatalf("UserForToken: %v", err)
	}
	if u.ID != 42 {
		t.Errorf("user ID = %d, want 42", u.ID)
	}
}

func TestSessionStop(t *testing.T) {
	users := &fakeUsers{byEmail: map[string]domain.User{}}
	sessions := &fakeSessions{byToken: map[string]fakeSession{}}
	svc := NewSessionService(sessions, users)

	token, _ := svc.Start(context.Background(), 1)
	if err := svc.Stop(context.Background(), token); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	_, err := svc.UserForToken(context.Background(), token)
	if !errors.Is(err, ports.ErrSessionNotFound) {
		t.Errorf("after Stop, UserForToken err = %v, want ErrSessionNotFound", err)
	}

	// Stopping again must not error (idempotent).
	if err := svc.Stop(context.Background(), token); err != nil {
		t.Errorf("second Stop errored: %v", err)
	}
}
