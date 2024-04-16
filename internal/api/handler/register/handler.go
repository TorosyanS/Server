package register

import (
	"encoding/json"
	"errors"
	"net/http"
	"test/internal/custom_errors"

	"test/internal/api/entity"
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

	err = h.authService.RegisterUser(body.Login, body.Password)
	if err != nil {
		if errors.Is(err, custom_errors.ErrUserAlreadyExists) {
			apiError, _ := json.Marshal(entity.ApiError{Message: err.Error()})
			w.WriteHeader(http.StatusConflict)
			w.Write(apiError)
			return
		}
		apiError, _ := json.Marshal(entity.ApiError{Message: "cannot authorize user"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	response, _ := json.Marshal(ResponseBody{Success: true})
	w.Write(response)
}
