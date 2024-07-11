package tokens_test

import (
	"testing"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/stretchr/testify/assert"
)

func TestAccess(t *testing.T) {
	tests := []struct {
		name   string
		access tokens.Access
		err    error
		secret []byte
	}{
		{
			name: "success",
			access: tokens.Access{
				Id:    1,
				Email: "email@email.com",
				Exp:   time.Now().Add(time.Hour).Unix(),
			},
			err:    nil,
			secret: []byte("secret"),
		},
		{
			name: "expired",
			access: tokens.Access{
				Id:    1,
				Email: "email@email.com",
				Exp:   0,
			},
			err:    tokens.ErrExpired,
			secret: []byte("secret"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got, _ := tc.access.SignSecret(tc.secret)

			access, err := got.ParseVerifySecret(tc.secret)
			if tc.err != nil {
				assert.EqualError(tt, err, tc.err.Error())
				return
			} else {
				assert.NoError(tt, err)
				assert.Equal(tt, tc.access, *access)
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	tests := []struct {
		name    string
		refresh tokens.Refresh
		err     error
		secret  []byte
	}{
		{
			name: "success",
			refresh: tokens.Refresh{
				Id:  1,
				Exp: time.Now().Add(time.Hour).Unix(),
			},
			err:    nil,
			secret: []byte("secret"),
		},
		{
			name: "expired",
			refresh: tokens.Refresh{
				Id:  1,
				Exp: 0,
			},
			err:    tokens.ErrExpired,
			secret: []byte("secret"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got, _ := tc.refresh.SignSecret(tc.secret)

			access, err := got.ParseVerifySecret(tc.secret)
			if tc.err != nil {
				assert.EqualError(tt, err, tc.err.Error())
				return
			} else {
				assert.NoError(tt, err)
				assert.Equal(tt, tc.refresh, *access)
			}
		})
	}
}
