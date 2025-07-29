package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/DIMO-Network/server-garage/pkg/monserver"
	"github.com/DIMO-Network/server-garage/pkg/runner"
	"github.com/DIMO-Network/shared/pkg/settings"
	"github.com/DIMO-Network/telemetry-api/internal/app"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", app.AppName).Logger()
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) == 40 {
				logger = logger.With().Str("commit", s.Value[:7]).Logger()
				break
			}
		}
	}
	zerolog.DefaultContextLogger = &logger

	mainCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		<-mainCtx.Done()
		logger.Info().Msg("Received signal, shutting down...")
		cancel()
	}()

	runnerGroup, runnerCtx := errgroup.WithContext(mainCtx)

	settingsFile := flag.String("settings", "settings.yaml", "settings file")
	flag.Parse()

	cfg, err := settings.LoadConfig[config.Settings](*settingsFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't load settings.")
	}

	application, err := app.New(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create application.")
	}

	defer application.Cleanup()

	// Create monitoring server with query recorder endpoint
	monSrv := monserver.NewMonitoringServer(&logger, cfg.EnablePprof)

	// Create custom monitoring mux to add query recorder endpoint
	monSrv.Handle("/queries", application.QueryRecorder.Handler()) // Add query recorder endpoint

	runner.RunHandler(runnerCtx, runnerGroup, monSrv, ":"+strconv.Itoa(cfg.MonPort))

	mux := http.NewServeMux()
	mux.Handle("/", app.LoggerMiddleware(app.PanicRecoveryMiddleware(playground.Handler("GraphQL playground", "/query"))))
	mux.Handle("/query", application.Handler)

	logger.Info().Msgf("Server started on port: %d", cfg.Port)
	runner.RunHandler(runnerCtx, runnerGroup, mux, ":"+strconv.Itoa(cfg.Port))

	err = runnerGroup.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Fatal().Err(err).Msg("Server shut down due to an error.")
	}
	logger.Info().Msg("Server shut down.")
}
