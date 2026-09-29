package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

const SessionLifetime = 7 * 24 * time.Hour // one week

type SessionService struct {
	Sessions ports.SessionStore
	Users    ports.UserRepository
}

func NewSessionService(sessions ports.SessionStore, users ports.UserRepository) *SessionService {
	return &SessionService{Sessions: sessions, Users: users}
}

func (s *SessionService) Start(ctx context.Context, userID int) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	expires := time.Now().Add(SessionLifetime)
	if err := s.Sessions.Create(ctx, token, userID, expires); err != nil {
		return "", err
	}
	return token, nil
}

func (s *SessionService) UserForToken(ctx context.Context, token string) (domain.User, error) {
	if token == "" {
		return domain.User{}, ports.ErrSessionNotFound
	}

	userID, err := s.Sessions.FindUserID(ctx, token)
	if err != nil {
		return domain.User{}, err
	}

	return s.Users.FindByID(ctx, userID)
}

func (s *SessionService) Stop(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.Sessions.Delete(ctx, token)
}

func newToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}
