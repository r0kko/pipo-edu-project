package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleGuard    Role = "guard"
	RoleResident Role = "resident"
)

type Claims struct {
	UserID string `json:"sub"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewTokenManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (tm *TokenManager) GenerateTokens(userID uuid.UUID, role Role) (access string, refresh string, err error) {
	now := time.Now()
	access, err = tm.signToken(userID, role, now.Add(tm.accessTTL), tm.accessSecret)
	if err != nil {
		return "", "", err
	}
	refresh, err = tm.signToken(userID, role, now.Add(tm.refreshTTL), tm.refreshSecret)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (tm *TokenManager) ParseAccess(token string) (*Claims, error) {
	return tm.parse(token, tm.accessSecret)
}

func (tm *TokenManager) ParseRefresh(token string) (*Claims, error) {
	return tm.parse(token, tm.refreshSecret)
}

func (tm *TokenManager) signToken(userID uuid.UUID, role Role, expiresAt time.Time, secret []byte) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (tm *TokenManager) parse(token string, secret []byte) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
