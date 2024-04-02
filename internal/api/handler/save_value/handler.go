package save_value

import (
	"encoding/json"
	"net/http"

	"test/internal/api/entity"
	"test/internal/polymorphism/storage"
)

type Handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {
	return &Handler{
		storage: storage,
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

	err = h.storage.SavePair(body.Key, body.Value)
	if err != nil {
		apiError, _ := json.Marshal(entity.ApiError{Message: "cannot save key-value"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(apiError)
		return
	}

	response, _ := json.Marshal(ResponseBody{Success: true})
	w.Write(response)
}
