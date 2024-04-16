package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func main() {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "telemetry-api").Logger()
	ns := os.Getenv("POD_NAMESPACE")
	if ns != "dev" {
		logger.Fatal().Msg("This version of the telemetry-api is only meant to be run on the dev. revert this commit to run on prod")
	}
	// create a flag for the settings file
	settingsFile := flag.String("settings", "settings.yaml", "settings file")
	flag.Parse()
	settings, err := shared.LoadConfig[config.Settings](*settingsFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't load settings.")
	}
	// create clickhouse connection
	_ = ctx

	repoLogger := logger.With().Str("component", "repository").Logger()
	baseRepo, err := repositories.NewRepository(&repoLogger, settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create repository.")
	}

	cfg := graph.Config{Resolvers: &graph.Resolver{Repository: baseRepo}}
	cfg.Directives.OneOf = func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
		// The directive on its own is advisory; everything is enforced inside of the resolver
		return next(ctx)
	}
	cfg.Directives.RequiresPrivilege = requiresPrivilegeCheck
	cfg.Directives.RequiresToken = requiresTokenCheck

	serveMonitoring(strconv.Itoa(settings.MonPort), &logger)

	server := handler.NewDefaultServer(graph.NewExecutableSchema(cfg))
	errLogger := logger.With().Str("component", "gql").Logger()
	server.SetErrorPresenter(errorHandler(errLogger))

	logger.Info().Str("jwksUrl", settings.TokenExchangeJWTKeySetURL).Str("issuerURL", settings.TokenExchangeIssuer).Str("vehicleAddr", settings.VehicleNFTAddress).Msg("Privileges enabled.")

	authMiddleware, err := NewJWTMiddleware(settings.TokenExchangeIssuer, settings.TokenExchangeJWTKeySetURL, settings.VehicleNFTAddress, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create JWT middleware.")
	}

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))

	authedHandler := authMiddleware.CheckJWT(AddClaimHandler(server, &logger))
	http.Handle("/query", authedHandler)

	logger.Info().Msgf("Server started on port: %d", settings.Port)

	logger.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil)).Msg("Server shut down.")
}

func serveMonitoring(port string, logger *zerolog.Logger) *fiber.App {
	logger.Info().Str("port", port).Msg("Starting monitoring web server.")

	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	monApp.Get("/", func(c *fiber.Ctx) error { return nil })
	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	go func() {
		if err := monApp.Listen(":" + port); err != nil {
			logger.Fatal().Err(err).Str("port", port).Msg("Failed to start monitoring web server.")
		}
	}()

	return monApp
}

// errorHandler is a custom error handler for gqlgen
func errorHandler(log zerolog.Logger) func(ctx context.Context, e error) *gqlerror.Error {
	return func(ctx context.Context, e error) *gqlerror.Error {
		var gqlErr *gqlerror.Error
		if errors.As(e, &gqlErr) {
			return gqlErr
		}
		log.Error().Err(e).Msg("Internal server error")
		return gqlerror.Errorf("internal server error")
	}
}
