package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime/debug"
	runtimepprof "runtime/pprof"
	"strconv"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/DIMO-Network/shared/pkg/settings"
	"github.com/DIMO-Network/telemetry-api/internal/app"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "telemetry-api").Logger()
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) == 40 {
				logger = logger.With().Str("commit", s.Value[:7]).Logger()
				break
			}
		}
	}
	zerolog.DefaultContextLogger = &logger

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

	serveMonitoring(strconv.Itoa(cfg.MonPort), &logger, cfg.EnablePprof)
	mux := http.NewServeMux()
	mux.Handle("/", loggerMiddleware(panicRecoveryMiddleware(playground.Handler("GraphQL playground", "/query"))))
	mux.Handle("/query", loggerMiddleware(panicRecoveryMiddleware(application.Handler)))

	logger.Info().Msgf("Server started on port: %d", cfg.Port)
	logger.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux)).Msg("Server shut down.")
}

func serveMonitoring(port string, logger *zerolog.Logger, enablePprof bool) *fiber.App {
	logger.Info().Str("port", port).Bool("pprof", enablePprof).Msg("Starting monitoring web server.")

	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Add panic recovery middleware
	monApp.Use(fiberrecover.New(fiberrecover.Config{
		EnableStackTrace: true,
	}))

	monApp.Get("/", func(c *fiber.Ctx) error { return nil })
	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Add pprof handlers if enabled
	if enablePprof {
		pprofGroup := monApp.Group("/debug/pprof")

		// Index page and base profiles
		pprofGroup.Get("/", adaptor.HTTPHandlerFunc(pprof.Index))
		pprofGroup.Get("/cmdline", adaptor.HTTPHandlerFunc(pprof.Cmdline))
		pprofGroup.Get("/profile", adaptor.HTTPHandlerFunc(pprof.Profile))
		pprofGroup.Get("/symbol", adaptor.HTTPHandlerFunc(pprof.Symbol))
		pprofGroup.Get("/trace", adaptor.HTTPHandlerFunc(pprof.Trace))

		// add specialized profiles
		profiles := runtimepprof.Profiles()
		for _, profile := range profiles {
			pprofGroup.Get("/"+profile.Name(), adaptor.HTTPHandler(pprof.Handler(profile.Name())))
		}

		logger.Info().Str("endpoint", "/debug/pprof").Msg("pprof profiling enabled on monitoring server")
	}

	go func() {
		if err := monApp.Listen(":" + port); err != nil {
			logger.Fatal().Err(err).Str("port", port).Msg("Failed to start monitoring web server.")
		}
	}()

	return monApp
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get source ip from request could be cloudflare proxy
		sourceIP := r.Header.Get("X-Forwarded-For")
		if sourceIP == "" {
			sourceIP = r.Header.Get("X-Real-IP")
		}
		if sourceIP == "" {
			sourceIP = r.RemoteAddr
		}
		loggerCtx := zerolog.Ctx(r.Context()).With().Str("method", r.Method).Str("path", r.URL.Path).Str("sourceIp", sourceIP).Logger().WithContext(r.Context())
		r = r.WithContext(loggerCtx)
		next.ServeHTTP(w, r)
	})
}

func panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "panic: %v\n%s\n", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
