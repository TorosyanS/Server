package get_value

import (
	"encoding/json"
	"errors"
	"net/http"

	"test/internal/api/entity"
	"test/internal/polymorphism/storage"
)

const queryParamKey = "key"

type Handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get(queryParamKey)
	if key == "" {
		apiError, _ := json.Marshal(entity.ApiError{Message: "incorrect body"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(apiError)
		return
	}

	value, err := h.storage.GetValue(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			apiError, _ := json.Marshal(entity.ApiError{Message: "not found"})
			w.WriteHeader(http.StatusNotFound)
			w.Write(apiError)
			return
		}

		apiError, _ := json.Marshal(entity.ApiError{Message: "cannot get value"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	response, _ := json.Marshal(ResponseBody{Value: value})
	w.Write(response)
}
