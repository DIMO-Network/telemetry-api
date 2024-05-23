package repositories

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/rs/zerolog"
)

var errInternal = errors.New("internal error")

// Repository is the base repository for all repositories.
type Repository struct {
	chService *ch.Service
	log       *zerolog.Logger
}

// NewRepository creates a new base repository.
// clientCAs is optional and can be nil.
func NewRepository(logger *zerolog.Logger, settings config.Settings, rootCAs *x509.CertPool) (*Repository, error) {
	chService, err := ch.NewService(settings, rootCAs)
	if err != nil {
		return nil, fmt.Errorf("failed to create ch service: %w", err)
	}
	return &Repository{
		chService: chService,
		log:       logger,
	}, nil
}

// GetSignal returns the aggregated signals for the given tokenID, interval, from, to and filter.
func (r *Repository) GetSignal(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*model.SignalAggregations, error) {
	if err := validateAggSigArgs(aggArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	signals, err := r.chService.GetAggregatedSignals(ctx, aggArgs)
	if err != nil {
		// Do not return the database erorr to the client, but log it.
		r.log.Error().Err(err).Msg("failed to query signals")
		return nil, errInternal
	}

	// combine signals with the same timestamp by iterating over all signals
	// if the timestamp differs from the previous signal, create a new SignalAggregations object
	var allAggs []*model.SignalAggregations
	var currAggs *model.SignalAggregations
	lastTS := time.Time{}
	for _, signal := range signals {
		if !lastTS.Equal(signal.Timestamp) {
			lastTS = signal.Timestamp
			currAggs = &model.SignalAggregations{
				Timestamp: signal.Timestamp,
			}
			allAggs = append(allAggs, currAggs)
		}

		model.SetAggregationField(currAggs, signal)
	}

	// add the aggregations to the last SignalAggregations object
	if currAggs != nil {
		allAggs = append(allAggs, currAggs)
	}
	return allAggs, nil
}

// GetSignalLatest returns the latest signals for the given tokenID and filter.
func (r *Repository) GetSignalLatest(ctx context.Context, latestArgs *model.LatestSignalsArgs) (*model.SignalCollection, error) {
	if err := validateSignalArgs(&latestArgs.SignalArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	signals, err := r.chService.GetLatestSignals(ctx, latestArgs)
	if err != nil {
		// Do not return the database erorr to the client, but log it.
		r.log.Error().Err(err).Msg("failed to query latest signals")
		return nil, errInternal
	}
	coll := &model.SignalCollection{}
	for _, signal := range signals {
		if signal.Name == model.LastSeenField {
			coll.LastSeen = &signal.Timestamp
			continue
		}
		model.SetCollectionField(coll, signal)
	}

	return coll, nil
}
