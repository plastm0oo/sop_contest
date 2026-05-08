package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/plastm0oo/sop_contest/internal/config"
	"github.com/plastm0oo/sop_contest/internal/middleware"
	"github.com/plastm0oo/sop_contest/internal/migrator"
	"github.com/plastm0oo/sop_contest/internal/service"
	deliveryhttp "github.com/plastm0oo/sop_contest/internal/service/delivery/http"
	"github.com/plastm0oo/sop_contest/internal/service/repository"
	"github.com/plastm0oo/sop_contest/internal/service/usecase"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Для локального запуска из VSCode / терминала
	// В Docker Compose env обычно уже подставляются снаружи.
	//if err := godotenv.Load(); err != nil {
	//	log.Println(".env file not loaded, continuing with system env")
	//}

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}
	defer db.Close()

	migrationCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrator.Run(migrationCtx, db, "migrations"); err != nil {
		log.Fatalf("migrations error: %v", err)
	}
	/*
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pingCancel()

		if err := db.PingContext(pingCtx); err != nil {
			log.Fatalf("db ping error: %v", err)
		}
	*/
	authCfg := service.AuthConfig{
		JWTSecret:            cfg.JWTSecret,
		AdminEmail:           cfg.AdminEmail,
		AccessTokenDuration:  cfg.AccessTokenDuration,
		RefreshTokenDuration: cfg.RefreshTokenDuration,
		BcryptCost:           cfg.BcryptCost,
	}

	repo := repository.New(db)
	uc := usecase.New(repo, authCfg)
	h := deliveryhttp.New(uc, cfg.JWTSecret)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	corsMiddleware := middleware.NewCORS(cfg.CORSAllowedOrigin)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitAttempts, cfg.RateLimitWindow)

	handlerChain := loggingMiddleware(
		corsMiddleware.Middleware(
			rateLimiter.Middleware(mux),
		),
	)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handlerChain,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("server started on :%s", cfg.Port)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}

		serverErrors <- nil
	}()

	select {
	case err := <-serverErrors:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	case <-ctx.Done():
		log.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped gracefully")
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		log.Printf(
			"method=%s path=%s status=%d duration=%s remote=%s",
			r.Method,
			r.URL.Path,
			recorder.status,
			time.Since(start).String(),
			clientIP(r),
		)
	})
}

func clientIP(r *http.Request) string {
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
