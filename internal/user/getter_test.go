package user

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

var (
	secret    = []byte("secret")
	validId   = 99
	validName = "luna"
)

func validToken() string {
	claims := jwt.MapClaims{
		"id":   validId,
		"name": validName,
		"exp":  time.Now().Add(60 * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, _ := token.SignedString(secret)

	return s
}

func tokenMissingId() string {
	claims := jwt.MapClaims{
		"name": validName,
		"exp":  time.Now().Add(60 * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, _ := token.SignedString(secret)

	return s
}

func tokenMissingUsername() string {
	claims := jwt.MapClaims{
		"id":  validId,
		"exp": time.Now().Add(60 * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, _ := token.SignedString(secret)

	return s
}

func invalidSignatureToken() string {
	claims := jwt.MapClaims{
		"id":   validId,
		"name": validName,
		"exp":  time.Now().Add(60 * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, _ := token.SignedString([]byte("wrong"))

	return s
}

func expiredToken() string {
	claims := jwt.MapClaims{
		"id":   validId,
		"name": validName,
		"exp":  time.Now().Add(-60 * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, _ := token.SignedString(secret)

	return s
}

func malformedToken() string {
	return "notatoken"
}

func TestJwtUserFromRequestGetter_GetUserFromRequest(t *testing.T) {
	getter := NewJwtUserFromRequestGetter(secret)

	tests := []struct {
		name          string
		token         string
		tokenHeader   bool
		expectedModel Model
		wantErr       bool
		expectedErr   string
	}{
		{
			name:          "valid token",
			token:         validToken(),
			tokenHeader:   true,
			expectedModel: Model{Id: validId, Username: validName},
			wantErr:       false,
		},
		{
			name:        "missing jwt-token header",
			token:       "",
			tokenHeader: false,
			wantErr:     true,
			expectedErr: "jwt-token header required",
		},
		{
			name:        "token missing id",
			token:       tokenMissingId(),
			tokenHeader: true,
			wantErr:     true,
			expectedErr: "token missing user id or username",
		},
		{
			name:        "token missing username",
			token:       tokenMissingUsername(),
			tokenHeader: true,
			wantErr:     true,
			expectedErr: "token missing user id or username",
		},
		{
			name:        "invalid signature",
			token:       invalidSignatureToken(),
			tokenHeader: true,
			wantErr:     true,
			expectedErr: "invalid token",
		},
		{
			name:        "expired token",
			token:       expiredToken(),
			tokenHeader: true,
			wantErr:     true,
			expectedErr: "invalid token",
		},
		{
			name:        "malformed token string",
			token:       malformedToken(),
			tokenHeader: true,
			wantErr:     true,
			expectedErr: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRequest("GET", "/", nil)

			if tt.tokenHeader {
				rr.Header.Set("jwt-token", tt.token)
			}

			got, err := getter.GetUserFromRequest(rr)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, Model{}, got)

				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedModel, got)
			}
		})
	}
}
