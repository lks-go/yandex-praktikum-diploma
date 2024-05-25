package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/lks-go/yandex-praktikum-diploma/internal/controller/handler"
)

func New() *app {
	return &app{}
}

type app struct {
}

func (a *app) Run(cfg Config) error {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx, stopApp := context.WithCancel(context.Background())
	defer stopApp()

	log := logrus.New()

	log.Info("running migrations")
	if err := RunMigrations(cfg.DatabaseDSN, "././internal/migrations"); err != nil {
		return fmt.Errorf("feiled to run migraions: %w", err)
	}

	h := handler.New()

	r := chi.NewRouter()
	r.Post("/api/user/register", h.RegisterUser)
	r.Post("/api/user/login", h.LoginUser)
	r.Post("/api/user/orders", h.SaveOrder)
	r.Get("/api/user/orders", h.Orders)
	r.Get("/api/user/balance", h.Balance)
	r.Post("/api/user/balance/withdraw", h.Withdraw)
	r.Get("/api/user/withdrawals", h.Withdrawals)

	addr := cfg.NetAddress.String()
	s := http.Server{
		Addr:    addr,
		Handler: r,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Info("starting http server")
		log.Infof("listen on %s\n", addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server failed: %w", err)
		}

		log.Info("http server stopped")

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		log.Info("shutdown http server")

		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), cfg.HttpServerShutdownTimeout)
		defer timeoutCancel()

		if err := s.Shutdown(timeoutCtx); err != nil {
			return fmt.Errorf("failed to shutdown http server: %w", err)
		}

		return nil
	})

	go func() {
		<-stopSignal
		stopApp()
	}()

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to wait errgroup: %w", err)
	}

	return nil
}
