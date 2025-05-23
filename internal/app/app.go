package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph"
	"github.com/DIMO-Network/telemetry-api/internal/limits"
	"github.com/DIMO-Network/telemetry-api/internal/metrics"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/attestation"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/DIMO-Network/telemetry-api/internal/service/fetchapi"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func init() {
	metrics.Register()
}

// App is the main application for the telemetry API.
type App struct {
	Handler http.Handler
	cleanup func()
}

// New creates a new application.
func New(settings config.Settings) (*App, error) {
	idService := identity.NewService(settings.IdentityAPIURL, settings.IdentityAPIReqTimeoutSeconds)
	chService, err := ch.NewService(settings)
	if err != nil {
		return nil, fmt.Errorf("couldn't create ClickHouse service: %w", err)
	}
	baseRepo, err := repositories.NewRepository(chService, settings.DeviceLastSeenBinHrs)
	if err != nil {
		return nil, fmt.Errorf("couldn't create base repository: %w", err)
	}
	vcRepo, err := newVinVCServiceFromSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("couldn't create VINVC repository: %w", err)
	}
	attRepo, err := newAttestationServiceFromSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create attestation repository: %w", err)
	}

	resolver := &graph.Resolver{
		Repository:      baseRepo,
		IdentityService: idService,
		VCRepo:          vcRepo,
		AttestationRepo: attRepo,
	}

	cfg := graph.Config{Resolvers: resolver}
	cfg.Directives.RequiresVehicleToken = auth.NewVehicleTokenCheck(settings.VehicleNFTAddress)
	cfg.Directives.RequiresManufacturerToken = auth.NewManufacturerTokenCheck(settings.ManufacturerNFTAddress, idService)
	cfg.Directives.RequiresAllOfPrivileges = auth.AllOfPrivilegeCheck
	cfg.Directives.RequiresOneOfPrivilege = auth.OneOfPrivilegeCheck
	cfg.Directives.IsSignal = noOp
	cfg.Directives.HasAggregation = noOp

	server := newDefaultServer(graph.NewExecutableSchema(cfg))

	authMiddleware, err := auth.NewJWTMiddleware(settings.TokenExchangeIssuer, settings.TokenExchangeJWTKeySetURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't create JWT middleware: %w", err)
	}

	limiter, err := limits.New(settings.MaxRequestDuration)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request time limit middleware: %w", err)
	}

	authedHandler := limiter.AddRequestTimeout(
		authMiddleware.CheckJWT(
			auth.AddClaimHandler(server, settings.VehicleNFTAddress, settings.ManufacturerNFTAddress),
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

func newAttestationServiceFromSettings(settings config.Settings) (*attestation.Repository, error) {
	fetchapiSvc := fetchapi.New(&settings)
	return attestation.New(fetchapiSvc, uint64(settings.ChainID), settings.VehicleNFTAddress), nil
}

func newVinVCServiceFromSettings(settings config.Settings) (*vc.Repository, error) {
	fetchapiSvc := fetchapi.New(&settings)
	return vc.New(fetchapiSvc, settings.VINVCDataVersion, settings.POMVCDataVersion, uint64(settings.ChainID), settings.VehicleNFTAddress), nil
}

func errorHandler(ctx context.Context, e error) *gqlerror.Error {
	var gqlErr *gqlerror.Error
	if errors.As(e, &gqlErr) {
		return gqlErr
	}
	zerolog.Ctx(ctx).Error().Err(e).Msg("Internal server error")
	return gqlerror.Errorf("internal server error")
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
	srv.Use(extension.FixedComplexityLimit(200))
	srv.Use(extension.Introspection{})
	srv.Use(metrics.Tracer{})

	srv.SetErrorPresenter(errorHandler)

	// add prometheus metrics
	srv.Use(metrics.Tracer{})

	return srv
}
