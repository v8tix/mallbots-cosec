package ms

import (
	"context"
	"database/sql"
	"github.com/v8tix/mallbots-cosec/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/v8tix/eda/waiter"
)

type Microservice interface {
	Config() config.AppConfig
	DB() *sql.DB
	JS() nats.JetStreamContext
	Logger() zerolog.Logger
	Mux() *chi.Mux
	RPC() *grpc.Server
	Waiter() waiter.Waiter
}

type Module interface {
	Startup(context.Context, Microservice) error
}
