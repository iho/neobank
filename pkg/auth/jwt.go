package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID    string `json:"uid"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

type JWT struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWT(secret string) *JWT {
	return &JWT{
		secret:     []byte(secret),
		accessTTL:  15 * time.Minute,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

func (j *JWT) Issue(userID string) (access, refresh string, err error) {
	access, err = j.sign(userID, TokenTypeAccess, j.accessTTL)
	if err != nil {
		return "", "", err
	}
	refresh, err = j.sign(userID, TokenTypeRefresh, j.refreshTTL)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (j *JWT) ValidateAccessToken(token string) (string, error) {
	return j.validate(token, TokenTypeAccess)
}

func (j *JWT) ValidateRefreshToken(token string) (string, error) {
	return j.validate(token, TokenTypeRefresh)
}

func (j *JWT) Refresh(refreshToken string) (access, newRefresh, userID string, err error) {
	userID, err = j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", "", err
	}
	access, newRefresh, err = j.Issue(userID)
	if err != nil {
		return "", "", "", err
	}
	return access, newRefresh, userID, nil
}

func (j *JWT) sign(userID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(j.secret)
}

func (j *JWT) validate(token, expectedType string) (string, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return "", fmt.Errorf("invalid token claims")
	}
	if claims.TokenType != expectedType {
		return "", fmt.Errorf("invalid token type")
	}
	if claims.UserID == "" {
		return "", fmt.Errorf("missing user id")
	}
	return claims.UserID, nil
}