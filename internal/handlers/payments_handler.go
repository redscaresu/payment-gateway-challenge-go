package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/go-chi/chi"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
)

type PaymentsHandler struct {
	storage *repository.PaymentsRepository
	domain  *domain.Domain
}

func NewPaymentsHandler(storage *repository.PaymentsRepository, domain *domain.Domain) *PaymentsHandler {
	return &PaymentsHandler{
		storage: storage,
		domain:  domain,
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
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var paymentRequest models.PostPaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
			log.Printf("Error decoding request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		domainResponse, err := ph.domain.PostPayment(&paymentRequest)
		if err != nil {
			log.Printf("Error processing payment: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set(contentTypeHeader, jsonContentType)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(domainResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

	}
}
