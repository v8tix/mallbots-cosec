package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	cosec "github.com/v8tix/mallbots-cosec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net/http"
	"os"
	"strings"

	"github.com/v8tix/eda/logger"
	"github.com/v8tix/eda/waiter"
	"github.com/v8tix/eda/web"
	"github.com/v8tix/mallbots-cosec/internal/config"
	"github.com/v8tix/mallbots-cosec/internal/ms"
)

func main() {
	var cfgDirFlag string
	var cfgFileFlag string
	var cfg config.AppConfig

	flag.StringVar(&cfgDirFlag, "d", "/home/v8tix/Public/projects/v8tix/microservices/environments/cloud/mallbots/cosec/dev", "The configuration directory")
	flag.StringVar(&cfgFileFlag, "f", "config", "The configuration file")
	flag.Parse()

	cfgFile := fmt.Sprintf("%s/%s", cfgDirFlag, cfgFileFlag)

	if err := run(cfgFile, &cfg); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run(configFile string, cfg *config.AppConfig) (err error) {
	err = config.InitConfig(configFile, cfg)
	if err != nil {
		return err
	}
	m := app{cfg: *cfg}

	// init infrastructure...
	// init db
	m.db, err = sql.Open("pgx", cfg.PG.Conn)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			return
		}
	}(m.db)
	// init nats & jetstream
	m.nc, err = nats.Connect(
		cfg.Nats.URL,
		nats.UserInfo(cfg.Nats.Username, cfg.Nats.Password),
		nats.Name(cfg.Nats.ClientName),
	)
	if err != nil {
		return err
	}
	defer m.nc.Close()
	m.js, err = initJetStream(cfg.Nats, m.nc)
	if err != nil {
		return err
	}
	m.logger = initLogger(cfg)
	m.rpc = initRPC(cfg.RPC)
	m.mux = initMux(cfg.Web)
	m.waiter = waiter.New(waiter.CatchSignals())

	// init modules
	m.modules = []ms.Module{
		&cosec.Module{},
	}

	if err = m.startupModules(); err != nil {
		return err
	}

	// Mount general web resources
	m.mux.Mount("/", http.FileServer(http.FS(web.WebUI)))

	fmt.Println("started mallbots application")
	defer fmt.Println("stopped mallbots application")

	m.waiter.Add(
		m.waitForWeb,
		m.waitForRPC,
		m.waitForStream,
	)

	// go func() {
	// 	for {
	// 		var mem runtime.MemStats
	// 		runtime.ReadMemStats(&mem)
	// 		m.logger.Debug().Msgf("Alloc = %v  TotalAlloc = %v  Sys = %v  NumGC = %v", mem.Alloc/1024, mem.TotalAlloc/1024, mem.Sys/1024, mem.NumGC)
	// 		time.Sleep(10 * time.Second)
	// 	}
	// }()

	return m.waiter.Wait()
}

func initLogger(cfg *config.AppConfig) zerolog.Logger {
	return logger.New(logger.LogConfig{
		Environment: cfg.Environment,
		LogLevel:    logger.Level(cfg.LogLevel),
	})
}

func initRPC(_ config.RPCConfig) *grpc.Server {
	server := grpc.NewServer()
	reflection.Register(server)

	return server
}

func initMux(_ config.WebConfig) *chi.Mux {
	return chi.NewMux()
}

func initJetStream(cfg config.NatsConfig, nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	streamCfg := nats.StreamConfig{
		Name:     cfg.Stream,
		Subjects: []string{fmt.Sprintf("%s.>", strings.ToLower(cfg.Stream))},
	}

	_, err = js.AddStream(&streamCfg)

	return js, err
}
