package user

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type JwtUserFromRequestGetter struct {
	secret []byte
}

func NewJwtUserFromRequestGetter(secret []byte) *JwtUserFromRequestGetter {
	return &JwtUserFromRequestGetter{
		secret: secret,
	}
}

func (g *JwtUserFromRequestGetter) GetUserFromRequest(r *http.Request) (Model, error) {
	tokenStr := strings.TrimSpace(r.Header.Get("jwt-token"))

	if tokenStr == "" {
		return Model{}, fmt.Errorf("jwt-token header required")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return g.secret, nil
	})

	if err != nil || !token.Valid {
		return Model{}, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return Model{}, fmt.Errorf("invalid token claims")
	}

	idFloat, hasId := claims["id"].(float64)
	username, hasName := claims["name"].(string)

	if !hasId || !hasName {
		return Model{}, fmt.Errorf("token missing user id or username")
	}

	return Model{
		Id:       int(idFloat),
		Username: username,
	}, nil
}
