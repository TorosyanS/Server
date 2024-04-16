package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"test/internal/api/entity"
	"test/internal/service/auth"
)

type AuthMiddleware struct {
	authService *auth.Service
}

func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (a *AuthMiddleware) CheckToken(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie("accessToken")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			apiError, _ := json.Marshal(entity.ApiError{Message: "access token not found"})
			w.Write(apiError)
			return
		}

		userLogin, err := a.authService.VerifyUser(accessToken.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			apiError, _ := json.Marshal(entity.ApiError{Message: err.Error()})
			w.Write(apiError)
			return
		}
		fmt.Printf("Пользователь %s - сделал запрос %s\n", userLogin, r.URL.Path)
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
