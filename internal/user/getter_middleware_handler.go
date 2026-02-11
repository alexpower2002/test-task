package user

import (
	"context"
	"net/http"
)

type UserFromRequestGetter interface {
	GetUserFromRequest(r *http.Request) (Model, error)
}

type UserGetterMiddleware struct {
	getter UserFromRequestGetter
	next   http.Handler
}

func NewUserGetterMiddleware(getter UserFromRequestGetter, next http.Handler) *UserGetterMiddleware {
	return &UserGetterMiddleware{
		getter: getter,
		next:   next,
	}
}

func (m *UserGetterMiddleware) Handle(w http.ResponseWriter, r *http.Request) {
	user, err := m.getter.GetUserFromRequest(r)

	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	ctx := context.WithValue(r.Context(), "userId", user.Id)

	m.next.ServeHTTP(w, r.WithContext(ctx))
}
