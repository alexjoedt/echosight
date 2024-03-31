package http

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/google/uuid"
)

func (s *Server) newSession(userID uuid.UUID) (*echosight.Session, error) {
	token, err := s.generateToken()
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(token))

	return &echosight.Session{
		Token:  token,
		Hash:   hash[:],
		Expiry: time.Now().Add(s.sessionLifetime),
		UserID: userID,
	}, nil
}

func (s *Server) generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
