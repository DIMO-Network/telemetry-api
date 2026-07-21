package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/DIMO-Network/server-garage/pkg/gql/metrics"
	"github.com/DIMO-Network/server-garage/pkg/mcpserver"
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
)

// App is the main application for the telemetry API.
type App struct {
	Handler       http.Handler
	MCPHandler    http.Handler
	QueryRecorder *queryRecorder.QueryRecorder
	cleanup       func()
}

// AppName is the name of the application.
var AppName = "telemetry-api"

// New creates a new application.
func New(settings config.Settings) (*App, error) {
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

	replayLogger, err := queryRecorder.NewReplayLogger(settings.RecordedDevelopers)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse recorded developer list: %w", err)
	}

	resolver := &graph.Resolver{
		BaseRepo:        baseRepo,
		VCRepo:          vcRepo,
		AttestationRepo: attRepo,
	}

	cfg := graph.Config{Resolvers: resolver}
	cfg.Directives.RequiresVehicleToken = auth.NewVehicleTokenCheck(settings.VehicleNFTAddress)
	cfg.Directives.RequiresAllOfPrivileges = auth.AllOfPrivilegeCheck
	cfg.Directives.RequiresOneOfPrivilege = auth.OneOfPrivilegeCheck
	cfg.Directives.IsSignal = noOp
	cfg.Directives.HasAggregation = noOp

	var costCalculator pricing.CostCalculator

	es := graph.NewExecutableSchema(cfg)
	dct := dtcmiddleware.NewDCT(ctClient, &costCalculator)

	server := newServer(es)
	configureGQLExtensions(server, dct, queryRec, replayLogger)

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
							auth.AddClaimHandler(server, settings.VehicleNFTAddress),
						),
					),
				),
			),
		),
	)

	mcpTools, err := graph.OverrideMCPTools(graph.MCPTools)
	if err != nil {
		return nil, fmt.Errorf("couldn't override MCP tools: %w", err)
	}

	// The MCP executor gets the same extension set as the HTTP GraphQL server
	// so tool calls are billed, recorded, and complexity-limited identically.
	mcpExec := mcpserver.NewGQLGenExecutor(es, func(e *executor.Executor) {
		configureGQLExtensions(e, dct, queryRec, replayLogger)
	})

	mcpHandler, err := mcpserver.New(mcpExec, "DIMO Telemetry", "0.1.0", "telemetry",
		mcpserver.WithTools(mcpTools),
		mcpserver.WithCondensedSchema(graph.CondensedSchema),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't create MCP handler: %w", err)
	}

	// MCP shares the same auth chain as GraphQL. Schema introspection (get_schema)
	// is reachable without a JWT because WithCredentialsOptional is true — this is
	// acceptable since the schema is not secret. All data-fetching tools enforce
	// @requiresVehicleToken at the GraphQL resolver level.
	mcpAuthHandler := PanicRecoveryMiddleware(
		LoggerMiddleware(
			limiter.AddRequestTimeout(
				authMiddleware.CheckJWT(
					authLoggerMiddleware(
						auth.AddClaimHandler(mcpHandler, settings.VehicleNFTAddress),
					),
				),
			),
		),
	)

	return &App{
		Handler:       serverHandler,
		MCPHandler:    mcpAuthHandler,
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
	// srv.SetQueryCache(graphql.NoCache[*ast.QueryDocument]{})

	return srv
}

// gqlExtensionTarget is satisfied by both *handler.Server and
// *executor.Executor, letting the HTTP GraphQL server and the MCP executor
// share one extension configuration.
type gqlExtensionTarget interface {
	Use(graphql.HandlerExtension)
	SetErrorPresenter(graphql.ErrorPresenterFunc)
}

// configureGQLExtensions applies the extension set shared by the HTTP GraphQL
// server and the MCP executor: complexity limit, introspection, metrics,
// error shaping, DCX credit tracking, and query recording. Keeping both paths
// on one list means MCP tool calls are billed and limited exactly like
// regular GraphQL requests.
func configureGQLExtensions(srv gqlExtensionTarget, dct *dtcmiddleware.DCT, queryRec *queryRecorder.QueryRecorder, replayLogger *queryRecorder.ReplayLogger) {
	srv.Use(extension.FixedComplexityLimit(100))
	srv.Use(extension.Introspection{})
	srv.Use(metrics.Tracer{})
	srv.SetErrorPresenter(errorPresenter)
	srv.Use(dct)
	srv.Use(queryRecorder.QueryRecordingExtension{Recorder: queryRec})
	if replayLogger != nil {
		srv.Use(replayLogger)
	}
}
