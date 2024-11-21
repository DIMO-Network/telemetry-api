package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/clickhouse-infra/pkg/connect"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph"
	"github.com/DIMO-Network/telemetry-api/internal/limits"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
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
	cfg.Directives.RequiresPrivileges = auth.PrivilegeCheck
	cfg.Directives.IsSignal = noOp
	cfg.Directives.HasAggregation = noOp
	cfg.Directives.OneOf = noOp

	server := handler.NewDefaultServer(graph.NewExecutableSchema(cfg))
	errLogger := logger.With().Str("component", "gql").Logger()
	server.SetErrorPresenter(errorHandler(errLogger))

	logger.Info().Str("jwksUrl", settings.TokenExchangeJWTKeySetURL).Str("issuerURL", settings.TokenExchangeIssuer).Str("vehicleAddr", settings.VehicleNFTAddress).Msg("Privileges enabled.")

	authMiddleware, err := auth.NewJWTMiddleware(settings.TokenExchangeIssuer, settings.TokenExchangeJWTKeySetURL, settings.VehicleNFTAddress, settings.ManufacturerNFTAddress, logger)
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
	chConfig := settings.CLickhouse
	chConfig.Database = settings.ClickhouseFileIndexDatabase
	chConn, err := connect.GetClickhouseConn(&chConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get clickhouse connection: %w", err)
	}
	s3Client, err := s3ClientFromSettings(&settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %w", err)
	}
	vinvcLogger := parentLogger.With().Str("component", "vinvc").Logger()
	if !common.IsHexAddress(settings.VehicleNFTAddress) {
		return nil, fmt.Errorf("invalid vehicle address: %s", settings.VehicleNFTAddress)
	}
	vehicleAddr := common.HexToAddress(settings.VehicleNFTAddress)
	return vc.New(chConn, s3Client, settings.VCBucket, settings.VINVCDataType, settings.POMVCDataType, uint64(settings.ChainID), vehicleAddr, &vinvcLogger), nil
}

// s3ClientFromSettings creates an S3 client from the given settings.
func s3ClientFromSettings(settings *config.Settings) (*s3.Client, error) {
	// Create an AWS session
	conf := aws.Config{
		Region: settings.S3AWSRegion,
		Credentials: credentials.NewStaticCredentialsProvider(
			settings.S3AWSAccessKeyID,
			settings.S3AWSSecretAccessKey,
			"",
		),
	}
	return s3.NewFromConfig(conf, func(o *s3.Options) {
		o.BaseEndpoint = settings.S3BaseEndpoint
	}), nil
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
