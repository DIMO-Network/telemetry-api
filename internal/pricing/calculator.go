// Package pricing provides cost calculation for GraphQL queries based on computational complexity.
// This package can be easily removed if the pricing model needs to be changed.
package pricing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/ast"
)

// PricingConfig contains configuration for the pricing calculator
type PricingConfig struct {
	BaseCosts              map[string]uint64 `yaml:"base_costs"`
	AggregationCosts       map[string]uint64 `yaml:"aggregation_costs"`
	TimeRangeCosts         map[string]uint64 `yaml:"time_range_costs"`
	IntervalCosts          map[string]uint64 `yaml:"interval_costs"`
	SignalCountCostDivisor uint64            `yaml:"signal_count_cost_divisor"`
}

// DefaultPricingConfig returns the default pricing configuration
func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		BaseCosts: map[string]uint64{
			"signals":          1, // Most expensive - time-series aggregations
			"signalsLatest":    1, // Medium - latest value lookups
			"events":           3, // Medium - event log queries
			"availableSignals": 1, // Cheap - metadata query
			"deviceActivity":   1, // Simple - device status lookup
			"vinVCLatest":      1, // Simple - credential lookup
			"pomVCLatest":      1, // Simple - credential lookup
		},
		AggregationCosts: map[string]uint64{
			// Float aggregations (from cheapest to most expensive)
			"MAX":   1,
			"MIN":   1,
			"FIRST": 1,
			"LAST":  1,
			"AVG":   2,
			"RAND":  2,
			"MED":   3, // Most expensive - requires sorting

			// String aggregations
			"TOP":    2,
			"UNIQUE": 4, // Most expensive - requires deduplication
		},
		TimeRangeCosts: map[string]uint64{
			"0-1h":  1,  // 0-1 hour
			"1-24h": 2,  // 1-24 hours
			"1-7d":  4,  // 1-7 days
			"1-30d": 8,  // 1-30 days
			"1-90d": 16, // 1-90 days
			"90d+":  32, // 90+ days
		},
		IntervalCosts: map[string]uint64{
			"sub-minute": 10, // Sub-minute (very expensive)
			"1-5min":     5,  // 1-5 minutes
			"5-60min":    3,  // 5-60 minutes
			"1-24h":      2,  // 1-24 hours
			"24h+":       1,  // 24+ hours
		},
		SignalCountCostDivisor: 5, // every 5 signals costs 1 credit
	}
}

// CostCalculator analyzes GraphQL queries and calculates credit costs
type CostCalculator struct {
	config *PricingConfig
}

// NewCostCalculator creates a new CostCalculator with the given configuration
func NewCostCalculator(config PricingConfig) *CostCalculator {
	return &CostCalculator{
		config: &config,
	}
}

// CalculateQueryCost analyzes a GraphQL operation and calculates the total cost
func (c *CostCalculator) CalculateQueryCost(ctx context.Context, operation *ast.OperationDefinition, variables map[string]interface{}) (uint64, error) {
	if c.config == nil {
		c.config = DefaultPricingConfig()
	}
	if operation == nil || operation.Operation != ast.Query {
		return 0, fmt.Errorf("cost calculator only supports query operations")
	}
	logger := zerolog.Ctx(ctx)

	var totalCost uint64 = 0

	for _, selection := range operation.SelectionSet {
		if field, ok := selection.(*ast.Field); ok {
			cost, err := c.calculateFieldCost(ctx, field, variables)
			if err != nil {
				return 0, fmt.Errorf("failed to calculate field cost: %w", err)
			}
			totalCost += cost
			logger.Debug().Str("field", field.Name).Uint64("cost", cost).Msg("Field cost calculated")
		}
	}

	// Ensure minimum cost of 1
	if totalCost == 0 {
		totalCost = 1
	}

	logger.Debug().Uint64("totalCost", totalCost).Msg("Total query cost calculated")
	return totalCost, nil
}

// calculateFieldCost calculates the cost for a specific GraphQL field
func (c *CostCalculator) calculateFieldCost(ctx context.Context, field *ast.Field, variables map[string]interface{}) (uint64, error) {
	fieldName := field.Name

	switch fieldName {
	case "signals":
		return c.calculateSignalsCost(ctx, field, variables)
	case "signalsLatest":
		return c.calculateSignalsLatestCost(field, variables)
	case "events":
		return c.calculateEventsCost(field, variables)
	default:
		return c.getBaseCost(fieldName)
	}
}

// calculateSignalsCost calculates cost for aggregated signals query
func (c *CostCalculator) calculateSignalsCost(ctx context.Context, field *ast.Field, variables map[string]interface{}) (uint64, error) {
	baseCost, err := c.getBaseCost("signals")
	if err != nil {
		return 0, fmt.Errorf("failed to get base cost: %w", err)
	}
	// Extract time range
	from, err := c.extractTimeArg(field, "from", variables)
	if err != nil {
		return 0, fmt.Errorf("failed to extract 'from' time: %w", err)
	}

	to, err := c.extractTimeArg(field, "to", variables)
	if err != nil {
		return 0, fmt.Errorf("failed to extract 'to' time: %w", err)
	}

	// Extract interval
	interval, err := c.extractStringArg(field, "interval", variables)
	if err != nil {
		return 0, fmt.Errorf("failed to extract 'interval': %w", err)
	}

	// Calculate component costs
	timeRangeCost := c.calculateTimeRangeCost(from, to)
	intervalCost, err := c.calculateIntervalCost(interval)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate interval cost: %w", err)
	}
	signalCount := countRequestedSignals(field)
	signalCountCost := c.calculateSignalCountCost(signalCount)
	aggregationCost := c.calculateAggregationCost(field)

	// Total cost is base cost multiplied by all factors
	totalCost := baseCost * timeRangeCost * intervalCost * signalCountCost * aggregationCost

	zerolog.Ctx(ctx).Debug().
		Uint64("baseCost", baseCost).
		Uint64("timeRangeCost", timeRangeCost).
		Uint64("intervalCost", intervalCost).
		Int("signalCount", signalCount).
		Uint64("signalCountCost", signalCountCost).
		Uint64("aggregationCost", aggregationCost).
		Uint64("totalCost", totalCost).
		Msg("Signals cost breakdown")

	return totalCost, nil
}

// calculateSignalsLatestCost calculates cost for latest signals query
func (c *CostCalculator) calculateSignalsLatestCost(field *ast.Field, variables map[string]interface{}) (uint64, error) {
	baseCost, err := c.getBaseCost("signalsLatest")
	if err != nil {
		return 0, fmt.Errorf("failed to get base cost: %w", err)
	}
	signalCount := countRequestedSignals(field)
	signalCountCost := c.calculateSignalCountCost(signalCount)

	totalCost := baseCost * signalCountCost
	return totalCost, nil
}

// calculateEventsCost calculates cost for events query
func (c *CostCalculator) calculateEventsCost(field *ast.Field, variables map[string]interface{}) (uint64, error) {
	baseCost, err := c.getBaseCost("events")
	if err != nil {
		return 0, fmt.Errorf("failed to get base cost: %w", err)
	}

	// Extract time range
	from, err := c.extractTimeArg(field, "from", variables)
	if err != nil {
		return 0, fmt.Errorf("failed to extract 'from' time: %w", err)
	}

	to, err := c.extractTimeArg(field, "to", variables)
	if err != nil {
		return 0, fmt.Errorf("failed to extract 'to' time: %w", err)
	}

	timeRangeCost := c.calculateTimeRangeCost(from, to)
	totalCost := baseCost * timeRangeCost

	return totalCost, nil
}

// calculateTimeRangeCost calculates cost multiplier based on time range duration
func (c *CostCalculator) calculateTimeRangeCost(from, to time.Time) uint64 {
	duration := to.Sub(from)
	hours := duration.Hours()

	switch {
	case hours <= 1:
		return c.config.TimeRangeCosts["0-1h"]
	case hours <= 24:
		return c.config.TimeRangeCosts["1-24h"]
	case hours <= 168:
		return c.config.TimeRangeCosts["1-7d"]
	case hours <= 720:
		return c.config.TimeRangeCosts["1-30d"]
	case hours <= 2160:
		return c.config.TimeRangeCosts["1-90d"]
	default:
		return c.config.TimeRangeCosts["90d+"]
	}
}

// calculateIntervalCost calculates cost multiplier based on aggregation interval granularity
func (c *CostCalculator) calculateIntervalCost(interval string) (uint64, error) {
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return 0, fmt.Errorf("failed to parse interval: %w", err)
	}

	switch {
	case duration < time.Minute:
		return c.config.IntervalCosts["sub-minute"], nil // Sub-minute (very expensive)
	case duration < 5*time.Minute:
		return c.config.IntervalCosts["1-5min"], nil // 1-5 minutes
	case duration < time.Hour:
		return c.config.IntervalCosts["5-60min"], nil // 5-60 minutes
	case duration < 24*time.Hour:
		return c.config.IntervalCosts["1-24h"], nil // 1-24 hours
	default:
		return c.config.IntervalCosts["24h+"], nil // 24+ hours
	}
}

// calculateSignalCountCost calculates cost multiplier based on number of signals requested
func (c *CostCalculator) calculateSignalCountCost(signalCount int) uint64 {
	// every N signals costs 1 credit (configurable)
	return uint64(signalCount/int(c.config.SignalCountCostDivisor)) + 1
}

// countRequestedSignals counts how many distinct signals are being requested
func countRequestedSignals(field *ast.Field) int {
	count := 0

	// Walk through the selection set to count signal fields
	for _, selection := range field.SelectionSet {
		if subField, ok := selection.(*ast.Field); ok {
			// Check if this field represents a signal (has @isSignal directive or matches signal patterns)
			if isSignalField(subField) {
				count++
			}
		}
	}

	return count
}

// isSignalField determines if a GraphQL field represents a signal
func isSignalField(field *ast.Field) bool {
	return field.Definition.Directives.ForName("isSignal") != nil
}

// calculateAggregationCost calculates cost multiplier based on aggregation complexity
func (c *CostCalculator) calculateAggregationCost(field *ast.Field) uint64 {
	maxCost := uint64(1)

	// Walk through selection set to find aggregation arguments
	for _, selection := range field.SelectionSet {
		if subField, ok := selection.(*ast.Field); ok {
			for _, arg := range subField.Arguments {
				if arg.Name == "agg" {
					if arg.Value.Kind == ast.EnumValue {
						enumValue := arg.Value.Raw
						if cost, exists := c.config.AggregationCosts[enumValue]; exists && cost > maxCost {
							maxCost = cost
						}
					}
				}
			}
		}
	}

	return maxCost
}

// getBaseCost returns the base cost for a field, defaulting to 1 if not found
func (c *CostCalculator) getBaseCost(fieldName string) (uint64, error) {
	if cost, exists := c.config.BaseCosts[fieldName]; exists {
		return cost, nil
	}
	return 0, fmt.Errorf("base cost not found for field: %s", fieldName)
}

// extractTimeArg extracts a time argument from GraphQL field arguments
func (c *CostCalculator) extractTimeArg(field *ast.Field, argName string, variables map[string]interface{}) (time.Time, error) {
	for _, arg := range field.Arguments {
		if arg.Name == argName {
			switch arg.Value.Kind {
			case ast.Variable:
				if varValue, exists := variables[arg.Value.Raw]; exists {
					if timeStr, ok := varValue.(string); ok {
						return time.Parse(time.RFC3339, timeStr)
					}
				}
			case ast.StringValue:
				return time.Parse(time.RFC3339, arg.Value.Raw)
			}
		}
	}
	return time.Time{}, fmt.Errorf("argument '%s' not found", argName)
}

// extractStringArg extracts a string argument from GraphQL field arguments
func (c *CostCalculator) extractStringArg(field *ast.Field, argName string, variables map[string]interface{}) (string, error) {
	for _, arg := range field.Arguments {
		if arg.Name == argName {
			switch arg.Value.Kind {
			case ast.Variable:
				if varValue, exists := variables[arg.Value.Raw]; exists {
					if strValue, ok := varValue.(string); ok {
						return strValue, nil
					}
				}
			case ast.StringValue:
				return arg.Value.Raw, nil
			}
		}
	}
	return "", fmt.Errorf("argument '%s' not found", argName)
}

// extractIntArg extracts an integer argument from GraphQL field arguments
func (c *CostCalculator) extractIntArg(field *ast.Field, argName string, variables map[string]interface{}) (int, error) {
	for _, arg := range field.Arguments {
		if arg.Name == argName {
			switch arg.Value.Kind {
			case ast.Variable:
				if varValue, exists := variables[arg.Value.Raw]; exists {
					if intValue, ok := varValue.(int); ok {
						return intValue, nil
					}
					if floatValue, ok := varValue.(float64); ok {
						return int(floatValue), nil
					}
					if strValue, ok := varValue.(string); ok {
						return strconv.Atoi(strValue)
					}
				}
			case ast.IntValue:
				return strconv.Atoi(arg.Value.Raw)
			}
		}
	}
	return 0, fmt.Errorf("argument '%s' not found", argName)
}
