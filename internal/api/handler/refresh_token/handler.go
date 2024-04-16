package refresh_token

import (
	"encoding/json"
	"errors"
	"net/http"

	"test/internal/api/entity"
	"test/internal/custom_errors"
	"test/internal/service/auth"
)

type Handler struct {
	authService *auth.Service
}

func NewHandler(authService *auth.Service) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		apiError, _ := json.Marshal(entity.ApiError{Message: "refresh token not found"})
		w.Write(apiError)
		return
	}

	tokens, err := h.authService.RefreshToken(refreshToken.Value)
	if err != nil {
		if errors.Is(err, custom_errors.ErrNotFound) {
			apiError, _ := json.Marshal(entity.ApiError{Message: "token not found in storage"})
			w.WriteHeader(http.StatusNotFound)
			w.Write(apiError)
			return
		}
		apiError, _ := json.Marshal(entity.ApiError{Message: "cannot refresh token"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	accessTokenCookie := http.Cookie{
		Name:     "accessToken",
		Value:    tokens.AccessToken,
		HttpOnly: true,
	}
	refreshTokenCookie := http.Cookie{
		Name:     "refreshToken",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
	}
	http.SetCookie(w, &accessTokenCookie)
	http.SetCookie(w, &refreshTokenCookie)

	response, _ := json.Marshal(ResponseBody{Status: true})
	w.Write(response)
}
