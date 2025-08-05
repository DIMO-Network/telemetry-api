// Package pricing provides cost calculation for GraphQL queries based on computational complexity.
// This package can be easily removed if the pricing model needs to be changed.
package pricing

import (
	"context"
	"fmt"
	"time"

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
		SignalCountCostDivisor: 1,
	}
}

// CostBreakdown provides detailed information about how cost was calculated
type CostBreakdown struct {
	Name          string          `json:"name"`
	Cost          uint64          `json:"cost"`
	Description   string          `json:"description,omitempty"`
	SubBreakdowns []CostBreakdown `json:"subBreakdowns,omitempty"`
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

// CalculateQueryCost analyzes a GraphQL operation and returns cost with detailed breakdown
func (c *CostCalculator) CalculateQueryCost(ctx context.Context, operation *ast.OperationDefinition, variables map[string]interface{}) (*CostBreakdown, error) {
	if c.config == nil {
		c.config = DefaultPricingConfig()
	}
	if operation == nil || operation.Operation != ast.Query {
		return nil, fmt.Errorf("cost calculator only supports query operations")
	}

	var fieldBreakdowns []CostBreakdown
	var totalCost uint64 = 0

	for _, selection := range operation.SelectionSet {
		if field, ok := selection.(*ast.Field); ok {
			fieldBreakdown, err := c.calculateFieldCostBreakdown(ctx, field, variables)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate field cost: %w", err)
			}
			fieldBreakdowns = append(fieldBreakdowns, *fieldBreakdown)
			totalCost += fieldBreakdown.Cost
		}
	}

	// Ensure minimum cost of 1
	if totalCost == 0 {
		totalCost = 1
	}

	breakdown := &CostBreakdown{
		Name:          operation.Name,
		Cost:          totalCost,
		Description:   fmt.Sprintf("Total query cost: %d credits", totalCost),
		SubBreakdowns: fieldBreakdowns,
	}

	return breakdown, nil
}

// calculateFieldCostBreakdown calculates the cost for a specific GraphQL field with detailed breakdown
func (c *CostCalculator) calculateFieldCostBreakdown(ctx context.Context, field *ast.Field, variables map[string]interface{}) (*CostBreakdown, error) {
	fieldName := field.Name

	switch fieldName {
	case "signals":
		return c.calculateSignalsCost(ctx, field, variables)
	case "signalsLatest":
		return c.calculateSignalsLatestCost(field, variables)
	case "events":
		return c.calculateEventsCost(field, variables)
	default:
		baseCost, err := c.getBaseCost(fieldName)
		if err != nil {
			return nil, err
		}
		return &CostBreakdown{
			Name:        field.Alias,
			Cost:        baseCost,
			Description: fmt.Sprintf("Base cost for %s field", fieldName),
		}, nil
	}
}

// calculateSignalsCost calculates cost for aggregated signals query with detailed breakdown
func (c *CostCalculator) calculateSignalsCost(ctx context.Context, field *ast.Field, variables map[string]interface{}) (*CostBreakdown, error) {
	baseCost, err := c.getBaseCost("signals")
	if err != nil {
		return nil, fmt.Errorf("failed to get base cost: %w", err)
	}

	// Extract time range
	from, err := c.extractTimeArg(field, "from", variables)
	if err != nil {
		return nil, fmt.Errorf("failed to extract 'from' time: %w", err)
	}

	to, err := c.extractTimeArg(field, "to", variables)
	if err != nil {
		return nil, fmt.Errorf("failed to extract 'to' time: %w", err)
	}

	// Extract interval
	interval, err := c.extractStringArg(field, "interval", variables)
	if err != nil {
		return nil, fmt.Errorf("failed to extract 'interval': %w", err)
	}

	// Calculate component costs
	timeRangeCost := c.calculateTimeRangeCost(from, to)
	intervalCost, err := c.calculateIntervalCost(interval)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate interval cost: %w", err)
	}
	aggregationCosts := c.calculateAggregationCost(field)

	// Total cost is base cost multiplied by all factors
	totalCost := baseCost * timeRangeCost.Cost * intervalCost.Cost * aggregationCosts.Cost

	// Create sub-breakdowns
	subBreakdowns := []CostBreakdown{
		{
			Name:        "base",
			Cost:        baseCost,
			Description: "Base cost for signals field",
		},
		timeRangeCost,
		intervalCost,
		aggregationCosts,
	}

	breakdown := &CostBreakdown{
		Name:          field.Alias,
		Cost:          totalCost,
		Description:   fmt.Sprintf("Signals query cost: %d × %d × %d × %d = %d", baseCost, timeRangeCost.Cost, intervalCost.Cost, aggregationCosts.Cost, totalCost),
		SubBreakdowns: subBreakdowns,
	}

	return breakdown, nil
}

// calculateSignalsLatestCost calculates cost for latest signals query with detailed breakdown
func (c *CostCalculator) calculateSignalsLatestCost(field *ast.Field, variables map[string]interface{}) (*CostBreakdown, error) {
	baseCost, err := c.getBaseCost("signalsLatest")
	if err != nil {
		return nil, fmt.Errorf("failed to get base cost: %w", err)
	}
	signalCount := countRequestedSignals(field)
	signalCountCost := c.calculateSignalCountCost(signalCount)

	totalCost := baseCost * signalCountCost.Cost

	// Create sub-breakdowns
	subBreakdowns := []CostBreakdown{
		{
			Name:        "base",
			Cost:        baseCost,
			Description: "Base cost for signalsLatest field",
		},
		signalCountCost,
	}

	breakdown := &CostBreakdown{
		Name:          field.Alias,
		Cost:          totalCost,
		Description:   fmt.Sprintf("SignalsLatest query cost: %d × %d = %d", baseCost, signalCountCost.Cost, totalCost),
		SubBreakdowns: subBreakdowns,
	}

	return breakdown, nil
}

// calculateEventsCost calculates cost for events query with detailed breakdown
func (c *CostCalculator) calculateEventsCost(field *ast.Field, variables map[string]interface{}) (*CostBreakdown, error) {
	baseCost, err := c.getBaseCost("events")
	if err != nil {
		return nil, fmt.Errorf("failed to get base cost: %w", err)
	}

	// Extract time range
	from, err := c.extractTimeArg(field, "from", variables)
	if err != nil {
		return nil, fmt.Errorf("failed to extract 'from' time: %w", err)
	}

	to, err := c.extractTimeArg(field, "to", variables)
	if err != nil {
		return nil, fmt.Errorf("failed to extract 'to' time: %w", err)
	}

	timeRangeCost := c.calculateTimeRangeCost(from, to)
	totalCost := baseCost * timeRangeCost.Cost

	// Create sub-breakdowns
	subBreakdowns := []CostBreakdown{
		{
			Name:        "base",
			Cost:        baseCost,
			Description: "Base cost for events field",
		},
		timeRangeCost,
	}

	breakdown := &CostBreakdown{
		Name:          field.Alias,
		Cost:          totalCost,
		Description:   fmt.Sprintf("Events query cost: %d × %d = %d", baseCost, timeRangeCost.Cost, totalCost),
		SubBreakdowns: subBreakdowns,
	}

	return breakdown, nil
}

// calculateTimeRangeCost calculates cost multiplier based on time range duration
func (c *CostCalculator) calculateTimeRangeCost(from, to time.Time) CostBreakdown {
	duration := to.Sub(from)
	hours := duration.Hours()
	cost := uint64(0)
	switch {
	case hours <= 1:
		cost = c.config.TimeRangeCosts["0-1h"]
	case hours <= 24:
		cost = c.config.TimeRangeCosts["1-24h"]
	case hours <= 168:
		cost = c.config.TimeRangeCosts["1-7d"]
	case hours <= 720:
		cost = c.config.TimeRangeCosts["1-30d"]
	case hours <= 2160:
		cost = c.config.TimeRangeCosts["1-90d"]
	default:
		cost = c.config.TimeRangeCosts["90d+"]
	}
	return CostBreakdown{
		Name: "timeRange",
		Cost: cost,
	}
}

// calculateIntervalCost calculates cost multiplier based on aggregation interval granularity
func (c *CostCalculator) calculateIntervalCost(interval string) (CostBreakdown, error) {
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return CostBreakdown{}, fmt.Errorf("failed to parse interval: %w", err)
	}
	cost := uint64(0)
	var description string
	switch {
	case duration < time.Minute:
		description = "Sub-minute"
		cost = c.config.IntervalCosts["sub-minute"] // Sub-minute (very expensive)
	case duration < 5*time.Minute:
		description = "1-5 minutes"
		cost = c.config.IntervalCosts["1-5min"] // 1-5 minutes
	case duration < time.Hour:
		description = "5-60 minutes"
		cost = c.config.IntervalCosts["5-60min"] // 5-60 minutes
	case duration < 24*time.Hour:
		description = "1-24 hours"
		cost = c.config.IntervalCosts["1-24h"] // 1-24 hours
	default:
		description = "24+ hours"
		cost = c.config.IntervalCosts["24h+"] // 24+ hours
	}
	return CostBreakdown{
		Name:        "interval",
		Cost:        cost,
		Description: description,
	}, nil
}

// calculateSignalCountCost calculates cost multiplier based on number of signals requested
func (c *CostCalculator) calculateSignalCountCost(signalCount int) CostBreakdown {
	// every N signals costs 1 credit (configurable)
	cost := uint64(signalCount/int(c.config.SignalCountCostDivisor)) + 1
	return CostBreakdown{
		Name:        "signalCount",
		Cost:        cost,
		Description: fmt.Sprintf("Signal count multiplier for %d signals", signalCount),
	}
}

// countRequestedSignals counts how many distinct signals are being requested
func countRequestedSignals(field *ast.Field) int {
	count := 0

	// Walk through the selection set to count signal fields
	for _, selection := range field.SelectionSet {
		if subField, ok := selection.(*ast.Field); ok && isSignalField(subField) {
			count++
		}
	}

	return count
}

// isSignalField determines if a GraphQL field represents a signal
func isSignalField(field *ast.Field) bool {
	return field.Definition.Directives.ForName("isSignal") != nil
}

// calculateAggregationCost calculates cost multiplier based on aggregation complexity
func (c *CostCalculator) calculateAggregationCost(field *ast.Field) CostBreakdown {
	costs := []CostBreakdown{}

	// Walk through selection set to find aggregation arguments
	for _, selection := range field.SelectionSet {
		if subField, ok := selection.(*ast.Field); ok && isSignalField(subField) {
			fieldCost := uint64(1)
			var description = fmt.Sprintf("Field cost for field %s", subField.Alias)
			for _, arg := range subField.Arguments {
				if arg.Name == "agg" {
					if arg.Value.Kind == ast.EnumValue {
						enumValue := arg.Value.Raw
						if aggCost, exists := c.config.AggregationCosts[enumValue]; exists {
							fieldCost *= aggCost
							description += fmt.Sprintf(" and aggregation %s", enumValue)
						}
					}
				}
			}
			costs = append(costs, CostBreakdown{
				Name:          subField.Alias,
				Cost:          fieldCost,
				Description:   description,
				SubBreakdowns: costs,
			})
		}
	}
	fieldCost := uint64(0)
	for _, cost := range costs {
		fieldCost += cost.Cost
	}
	return CostBreakdown{
		Name:          field.Alias,
		Cost:          fieldCost,
		Description:   "costs for aggregated fields",
		SubBreakdowns: costs,
	}
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
