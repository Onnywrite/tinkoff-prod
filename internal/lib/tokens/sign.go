package tokens

import (
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration = 0
	RefreshTTL    time.Duration = 0
)

type Access struct {
	Id    uint64
	Email string
	Exp   int64
}

type Refresh struct {
	Id       uint64
	Rotation uint64
	Exp      int64
}

func (a *Access) Sign() (AccessString, error) {
	return a.SignSecret(AccessSecret)
}

func (a *Access) SignSecret(secret []byte) (AccessString, error) {
	if AccessTTL != 0 {
		a.Exp = time.Now().Add(AccessTTL).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    a.Id,
		"email": a.Email,
		"exp":   a.Exp,
	})

	tknstr, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return AccessString(tknstr), nil
}

func (r *Refresh) Sign() (RefreshString, error) {
	return r.SignSecret(RefreshSecret)
}

func (r *Refresh) SignSecret(secret []byte) (RefreshString, error) {
	if AccessTTL != 0 {
		r.Exp = time.Now().Add(RefreshTTL).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  r.Id,
		"exp": r.Exp,
		"rtr": r.Rotation,
	})

	tknstr, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return RefreshString(tknstr), nil
}
