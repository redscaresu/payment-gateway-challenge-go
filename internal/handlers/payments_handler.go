package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/go-chi/chi"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
)

type PaymentsHandler struct {
	storage *repository.PaymentsRepository
}

func NewPaymentsHandler(storage *repository.PaymentsRepository) *PaymentsHandler {
	return &PaymentsHandler{
		storage: storage,
	}
}

// GetHandler returns an http.HandlerFunc that handles HTTP GET requests.
// It retrieves a payment record by its ID from the storage.
// The ID is expected to be part of the URL.
func (h *PaymentsHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		payment := h.storage.GetPayment(id)

		if payment == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set(contentTypeHeader, jsonContentType)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(payment); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (ph *PaymentsHandler) PostHandler() http.HandlerFunc {
	//TODO
	return nil
}
