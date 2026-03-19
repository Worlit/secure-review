package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/secure-review/internal/domain"
)

var _ domain.TokenGenerator = (*JWTTokenGenerator)(nil)

// JWTTokenGenerator implements TokenGenerator using JWT
type JWTTokenGenerator struct {
	secret        string
	expiry        time.Duration
	refreshExpiry time.Duration
}

// NewJWTTokenGenerator creates a new JWTTokenGenerator
func NewJWTTokenGenerator(secret string, expiry, refreshExpiry time.Duration) *JWTTokenGenerator {
	return &JWTTokenGenerator{
		secret:        secret,
		expiry:        expiry,
		refreshExpiry: refreshExpiry,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for a user
func (g *JWTTokenGenerator) GenerateToken(userID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secret))
}

// ValidateToken validates a JWT token and returns the user ID
func (g *JWTTokenGenerator) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(g.secret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	return userID, nil
}

// GenerateRefreshToken generates a refresh token
func (g *JWTTokenGenerator) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secret))
}
