package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/server-garage/pkg/gql/metrics"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/dtcmiddleware"
	"github.com/DIMO-Network/telemetry-api/internal/graph"
	"github.com/DIMO-Network/telemetry-api/internal/limits"
	"github.com/DIMO-Network/telemetry-api/internal/pricing"
	"github.com/DIMO-Network/telemetry-api/internal/queryRecorder"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/attestation"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/DIMO-Network/telemetry-api/internal/service/credittracker"
	"github.com/DIMO-Network/telemetry-api/internal/service/fetchapi"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity"
)

// App is the main application for the telemetry API.
type App struct {
	Handler       http.Handler
	QueryRecorder *queryRecorder.QueryRecorder
	cleanup       func()
}

// AppName is the name of the application.
var AppName = "telemetry-api"

// New creates a new application.
func New(settings config.Settings) (*App, error) {
	idService := identity.NewService(settings.IdentityAPIURL, settings.IdentityAPIReqTimeoutSeconds)
	chService, err := ch.NewService(settings)
	if err != nil {
		return nil, fmt.Errorf("couldn't create ClickHouse service: %w", err)
	}
	baseRepo, err := repositories.NewRepository(chService, settings)
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

	ctClient, err := credittracker.NewClient(&settings, AppName)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit tracker client: %w", err)
	}

	// Create query recorder
	queryRec := queryRecorder.New()

	resolver := &graph.Resolver{
		BaseRepo:        baseRepo,
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

	var costCalculator pricing.CostCalculator

	server := newServer(graph.NewExecutableSchema(cfg))
	server.Use(dtcmiddleware.NewDCT(ctClient, &costCalculator))

	// Add query recording middleware
	server.Use(queryRecorder.QueryRecordingExtension{Recorder: queryRec})

	authMiddleware, err := auth.NewJWTMiddleware(settings.TokenExchangeIssuer, settings.TokenExchangeJWTKeySetURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't create JWT middleware: %w", err)
	}

	limiter, err := limits.New(settings.MaxRequestDuration)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request time limit middleware: %w", err)
	}

	serverHandler := PanicRecoveryMiddleware(
		LoggerMiddleware(
			limiter.AddRequestTimeout(
				authMiddleware.CheckJWT(
					authLoggerMiddleware(
						dtcmiddleware.EstimateCostHeaderMiddleware(
							auth.AddClaimHandler(server, settings.VehicleNFTAddress, settings.ManufacturerNFTAddress),
						),
					),
				),
			),
		),
	)

	return &App{
		Handler:       serverHandler,
		QueryRecorder: queryRec,
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
	return vc.New(fetchapiSvc, settings), nil
}

func newServer(es graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(es)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.Use(extension.FixedComplexityLimit(100))
	srv.Use(extension.Introspection{})
	srv.Use(metrics.Tracer{})
	// srv.SetQueryCache(graphql.NoCache[*ast.QueryDocument]{})
	srv.SetErrorPresenter(errorhandler.ErrorPresenter)

	return srv
}
