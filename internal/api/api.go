package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"
)

const (
	bankURL = "http://localhost:8080"
)

type Api struct {
	router             *chi.Mux
	paymentsRepo       *repository.PaymentsRepository
	domain             *domain.Domain
	PostPaymentService *domain.PaymentServiceImpl
}

func New() *Api {
	a := &Api{}
	repo := repository.NewPaymentsRepository()
	a.paymentsRepo = repo
	client := client.NewClient(bankURL, 5*time.Second)
	postPaymentService := domain.NewPaymentServiceImpl(repo, client)
	a.domain = domain.NewDomain(repo, client, postPaymentService)
	a.setupRouter()

	return a
}

func (a *Api) Run(ctx context.Context, addr string) error {
	httpServer := &http.Server{
		Addr:        addr,
		Handler:     a.router,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()
		fmt.Printf("shutting down HTTP server\n")
		return httpServer.Shutdown(ctx)
	})

	g.Go(func() error {
		fmt.Printf("starting HTTP server on %s\n", addr)
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	return g.Wait()
}

func (a *Api) setupRouter() {
	a.router = chi.NewRouter()
	a.router.Use(middleware.Logger)

	a.router.Get("/ping", a.PingHandler())
	a.router.Get("/swagger/*", a.SwaggerHandler())

	a.router.Get("/api/payments/{id}", a.GetPaymentHandler())
	a.router.Post("/api/payments", a.PostPaymentHandler())
}
