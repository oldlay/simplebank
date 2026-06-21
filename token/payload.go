package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Different types of error returned by the VerifyToken function
var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token has invalid")
)

type TokenType byte

const (
	TokenTypeAccessToken  = 1
	TokenTypeRefreshToken = 2
)

// Payload contains the payload data of the token
type Payload struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Type     TokenType `json:"type"`
	Role     string    `json:"role"`
	IssueAt  time.Time `json:"issued_at"`
	ExpireAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token playload with specific a name and duration
func NewPayload(username string, duration time.Duration, role string, tokenType TokenType) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:       tokenID,
		Username: username,
		Type:     tokenType,
		Role:     role,
		IssueAt:  time.Now(),
		ExpireAt: time.Now().Add(duration),
	}
	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpireAt) {
		return ErrExpiredToken
	}
	return nil
}

func (payload *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{
		Time: payload.ExpireAt,
	}, nil
}
func (payload *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{
		Time: payload.IssueAt,
	}, nil
}
func (payload *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{
		Time: payload.IssueAt,
	}, nil
}
func (payload *Payload) GetIssuer() (string, error) {
	return "", nil
}
func (payload *Payload) GetSubject() (string, error) {
	return "", nil
}
func (payload *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}
