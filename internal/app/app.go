package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/dtcmiddleware"
	"github.com/DIMO-Network/telemetry-api/internal/graph"
	"github.com/DIMO-Network/telemetry-api/internal/limits"
	"github.com/DIMO-Network/telemetry-api/internal/metrics"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/attestation"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/DIMO-Network/telemetry-api/internal/service/credittracker"
	"github.com/DIMO-Network/telemetry-api/internal/service/fetchapi"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity"
	"github.com/DIMO-Network/telemetry-api/pkg/errorhandler"
	"github.com/rs/zerolog"
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

	ctClient, err := credittracker.NewClient(&settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit tracker client: %w", err)
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

	server := newServer(graph.NewExecutableSchema(cfg))
	server.Use(dtcmiddleware.NewDCT(ctClient))

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
						auth.AddClaimHandler(server, settings.VehicleNFTAddress, settings.ManufacturerNFTAddress),
					),
				),
			),
		),
	)

	return &App{
		Handler: serverHandler,
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
	srv.SetErrorPresenter(errorhandler.ErrorPresenter)

	return srv
}

// LoggerMiddleware adds the source IP to the logger.
func LoggerMiddleware(next http.Handler) http.Handler {
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

// authLoggerMiddleware adds the authenticated user to the logger
func authLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validateClaims, ok := auth.GetValidatedClaims(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		loggerCtx := zerolog.Ctx(r.Context()).With().Str("jwtSubject", validateClaims.RegisteredClaims.Subject).Logger()
		r = r.WithContext(loggerCtx.WithContext(r.Context()))
		next.ServeHTTP(w, r)
	})
}

// PanicRecoveryMiddleware recovers from panics and logs them.
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
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
