package auth

import (
	"errors"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var ErrInvalidTokenType = errors.New("invalid token type")

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type JWTAuth struct {
	tokenAuth       *jwtauth.JWTAuth
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTAuth(secret string, accessTTL, refreshTTL time.Duration) *JWTAuth {
	return &JWTAuth{
		tokenAuth:       jwtauth.New("HS256", []byte(secret), nil),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (j *JWTAuth) GenerateToken(userID int) (string, error) {
	claims := map[string]interface{}{
		"user_id": userID,
		"type":    "access",
	}

	jwtauth.SetExpiryIn(claims, j.accessTokenTTL)
	_, tokenString, err := j.tokenAuth.Encode(claims)

	return tokenString, err
}

func (j *JWTAuth) GenerateRefreshToken(userID int) (string, error) {
	claims := map[string]interface{}{
		"user_id": userID,
		"type":    "refresh",
	}

	jwtauth.SetExpiryIn(claims, j.refreshTokenTTL)
	_, tokenString, err := j.tokenAuth.Encode(claims)

	return tokenString, err
}

func (j *JWTAuth) GenerateTokenPair(userID int) (TokenPair, error) {
	accessToken, err := j.GenerateToken(userID)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := j.GenerateRefreshToken(userID)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (j *JWTAuth) ValidateRefreshToken(tokenString string) (int, error) {
	token, err := j.tokenAuth.Decode(tokenString)
	if err != nil {
		return 0, err
	}

	if err := jwt.Validate(token); err != nil {
		return 0, err
	}

	tokenType, ok := token.Get("type")
	if !ok || tokenType != "refresh" {
		return 0, ErrInvalidTokenType
	}

	userID, ok := token.Get("user_id")
	if !ok {
		return 0, errors.New("missing user_id claim")
	}

	uid, ok := userID.(float64)
	if !ok {
		return 0, errors.New("invalid user_id claim")
	}

	return int(uid), nil
}

func (j *JWTAuth) GetTokenAuth() *jwtauth.JWTAuth {
	return j.tokenAuth
}
