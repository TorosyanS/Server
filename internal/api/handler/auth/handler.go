package auth

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
	body := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		apiError, _ := json.Marshal(entity.ApiError{Message: "incorrect body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiError)
		return
	}

	tokens, err := h.authService.AuthUser(body.Login, body.Password)
	if err != nil {
		if errors.Is(err, custom_errors.ErrNotFound) {
			apiError, _ := json.Marshal(entity.ApiError{Message: "cannot find user"})
			w.WriteHeader(http.StatusNotFound)
			w.Write(apiError)
			return
		}
		if errors.Is(err, custom_errors.ErrIncorrectPassword) {
			apiError, _ := json.Marshal(entity.ApiError{Message: "incorrect password"})
			w.WriteHeader(http.StatusForbidden)
			w.Write(apiError)
			return
		}
		apiError, _ := json.Marshal(entity.ApiError{Message: "cannot authorize user"})
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
