package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/lks-go/yandex-praktikum-diploma/internal/controller/handler"
	"github.com/lks-go/yandex-praktikum-diploma/internal/controller/storage"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service/auth"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	log.Info("setup db")
	pool, err := setupDB(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("failed to setup DB: %w", err)
	}

	tokenBuilder := auth.New(&auth.Config{})

	store := storage.New(pool)

	serviceDeps := service.Deps{
		UserStorage:  store,
		TokenBuilder: tokenBuilder,
	}
	service := service.New(&service.Config{}, &serviceDeps)

	h := handler.New(log, service)

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

func setupDB(dsn string) (*sql.DB, error) {
	pool, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := pool.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database after connect: %w", err)
	}

	return pool, nil
}
