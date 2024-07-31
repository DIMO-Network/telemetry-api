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
	"github.com/DIMO-Network/clickhouse-infra/pkg/connect"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph"
	"github.com/DIMO-Network/telemetry-api/internal/limits"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vinvc"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func main() {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "telemetry-api").Logger()
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
	chService, err := ch.NewService(settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create ClickHouse service.")
	}
	baseRepo := repositories.NewRepository(&repoLogger, chService)
	vinvcRepo, err := newVinVCServiceFromSettings(settings, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create VINVC repository.")
	}
	resolver := &graph.Resolver{
		Repository: baseRepo,
		VINVCRepo:  vinvcRepo,
	}

	vehCheck := auth.VehicleTokenChecker{
		ContractAddr: common.HexToAddress(settings.VehicleNFTAddress),
	}

	cfg := graph.Config{Resolvers: resolver}
	cfg.Directives.RequiresVehicleToken = vehCheck.Check
	cfg.Directives.RequiresVehiclePrivilege = auth.PrivilegeCheck
	cfg.Directives.IsSignal = noOp
	cfg.Directives.HasAggregation = noOp

	serveMonitoring(strconv.Itoa(settings.MonPort), &logger)

	server := handler.NewDefaultServer(graph.NewExecutableSchema(cfg))
	errLogger := logger.With().Str("component", "gql").Logger()
	server.SetErrorPresenter(errorHandler(errLogger))

	logger.Info().Str("jwksUrl", settings.TokenExchangeJWTKeySetURL).Str("issuerURL", settings.TokenExchangeIssuer).Str("vehicleAddr", settings.VehicleNFTAddress).Msg("Privileges enabled.")

	authMiddleware, err := auth.NewJWTMiddleware(settings.TokenExchangeIssuer, settings.TokenExchangeJWTKeySetURL, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create JWT middleware.")
	}

	limiter, err := limits.New(settings.MaxRequestDuration)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create request time limit middleware.")
	}

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))

	authedHandler := limiter.AddRequestTimeout(
		authMiddleware.CheckJWT(
			auth.AddClaimHandler(server, &logger, settings.VehicleNFTAddress, settings.ManufacturerNFTAddress),
		),
	)
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

func noOp(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	return next(ctx)
}

func newVinVCServiceFromSettings(settings config.Settings, parentLogger *zerolog.Logger) (*vinvc.Repository, error) {
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
	return vinvc.New(chConn, s3Client, settings.VINVCBucket, settings.VINVCDataType, &vinvcLogger), nil
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
	return s3.NewFromConfig(conf), nil
}
