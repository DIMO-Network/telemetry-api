package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph"
	"github.com/DIMO-Network/telemetry-api/internal/limits"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/DIMO-Network/telemetry-api/internal/service/fetchapi"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// App is the main application for the telemetry API.
type App struct {
	Handler http.Handler
	cleanup func()
}

// New creates a new application.
func New(settings config.Settings, logger *zerolog.Logger) (*App, error) {
	idService := identity.NewService(settings.IdentityAPIURL, settings.IdentityAPIReqTimeoutSeconds)
	repoLogger := logger.With().Str("component", "repository").Logger()
	chService, err := ch.NewService(settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create ClickHouse service.")
	}
	baseRepo, err := repositories.NewRepository(&repoLogger, chService, settings.DeviceLastSeenBinHrs)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create base repository.")
	}
	vcRepo, err := newVinVCServiceFromSettings(settings, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create VINVC repository.")
	}
	resolver := &graph.Resolver{
		Repository:      baseRepo,
		IdentityService: idService,
		VCRepo:          vcRepo,
	}

	cfg := graph.Config{Resolvers: resolver}
	cfg.Directives.RequiresVehicleToken = auth.NewVehicleTokenCheck(settings.VehicleNFTAddress)
	cfg.Directives.RequiresManufacturerToken = auth.NewManufacturerTokenCheck(settings.ManufacturerNFTAddress, idService)
	cfg.Directives.RequiresAllOfPrivileges = auth.AllOfPrivilegeCheck
	cfg.Directives.RequiresOneOfPrivilege = auth.OneOfPrivilegeCheck
	cfg.Directives.IsSignal = noOp
	cfg.Directives.HasAggregation = noOp

	server := newDefaultServer(graph.NewExecutableSchema(cfg))
	errLogger := logger.With().Str("component", "gql").Logger()
	server.SetErrorPresenter(errorHandler(errLogger))

	logger.Info().Str("jwksUrl", settings.TokenExchangeJWTKeySetURL).Str("issuerURL", settings.TokenExchangeIssuer).Str("vehicleAddr", settings.VehicleNFTAddress.Hex()).Msg("Privileges enabled.")

	authMiddleware, err := auth.NewJWTMiddleware(settings.TokenExchangeIssuer, settings.TokenExchangeJWTKeySetURL, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create JWT middleware.")
	}

	limiter, err := limits.New(settings.MaxRequestDuration)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create request time limit middleware.")
	}

	authedHandler := limiter.AddRequestTimeout(
		authMiddleware.CheckJWT(
			auth.AddClaimHandler(server, logger, settings.VehicleNFTAddress, settings.ManufacturerNFTAddress),
		),
	)

	return &App{
		Handler: authedHandler,
		cleanup: func() {
			// TODO add cleanup logic for closing connections
		},
	}, nil
}

func (a *App) Cleanup() {
	if a.cleanup != nil {
		a.cleanup()
	}
}

func noOp(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	return next(ctx)
}

func newVinVCServiceFromSettings(settings config.Settings, parentLogger *zerolog.Logger) (*vc.Repository, error) {
	fetchapiSvc := fetchapi.New(&settings)
	vinvcLogger := parentLogger.With().Str("component", "vinvc").Logger()
	return vc.New(fetchapiSvc, settings.VCBucket, settings.VINVCDataType, settings.POMVCDataType, uint64(settings.ChainID), settings.VehicleNFTAddress, &vinvcLogger), nil
}

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

func newDefaultServer(es graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(es)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return srv
}
