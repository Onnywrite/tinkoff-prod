package tokens

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type AccessString string

type RefreshString string

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidPayload          = errors.New("invalid payload")
	ErrExpired                 = errors.New("token has expired")
	ErrUnknown                 = errors.New("unknown")
)

func (a *AccessString) ParseVerify() (*Access, error) {
	return a.ParseVerifySecret(AccessSecret)
}

func (a *AccessString) ParseVerifySecret(secret []byte) (*Access, error) {
	parser := jwt.Parser{
		SkipClaimsValidation: true,
		ValidMethods:         []string{"HS256"},
	}
	token, err := parser.Parse(string(*a), func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, ErrUnexpectedSigningMethod
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, ok := claims["exp"].(float64)
		if !ok || exp < float64(time.Now().Unix()) {
			return nil, ErrExpired
		}

		id, ok := claims["id"].(float64)
		if !ok {
			return nil, ErrInvalidPayload
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, ErrInvalidPayload
		}

		return &Access{
			Id:    uint64(id),
			Email: email,
			Exp:   int64(exp),
		}, nil
	}

	return nil, ErrUnknown
}

func (a *RefreshString) ParseVerify() (*Refresh, error) {
	return a.ParseVerifySecret(RefreshSecret)
}

func (a *RefreshString) ParseVerifySecret(secret []byte) (*Refresh, error) {
	parser := jwt.Parser{
		SkipClaimsValidation: true,
		ValidMethods:         []string{"HS256"},
	}
	token, err := parser.Parse(string(*a), func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, ErrUnexpectedSigningMethod
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, ok := claims["exp"].(float64)
		if !ok || exp < float64(time.Now().Unix()) {
			return nil, ErrExpired
		}

		id, ok := claims["id"].(float64)
		if !ok {
			return nil, ErrInvalidPayload
		}

		return &Refresh{
			Id:  uint64(id),
			Exp: int64(exp),
		}, nil
	}

	return nil, ErrUnknown
}
